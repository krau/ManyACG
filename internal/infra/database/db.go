package database

import (
	"context"
	"sync"

	"github.com/krau/ManyACG/internal/infra/config/runtimecfg"
	"github.com/krau/ManyACG/internal/model/entity"
	"github.com/krau/ManyACG/pkg/log"
	_ "github.com/ncruces/go-sqlite3/embed"
	"github.com/ncruces/go-sqlite3/gormlite"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	defaultDB         *DB
	initOnce          sync.Once
	ErrRecordNotFound = gorm.ErrRecordNotFound
)

type DB struct {
	db *gorm.DB
}

func Default() *DB {
	if defaultDB == nil {
		// initOnce.Do(func() {
		// 	okCh := make(chan struct{})
		// 	initDB(context.Background(), okCh)
		// 	select {
		// 	case <-context.Background().Done():
		// 		log.Fatal("Database initialization canceled")
		// 	case <-okCh:
		// 	}
		// })
		log.Fatal("database not initialized, please call Init() first")
	}
	return defaultDB
}

func Init(ctx context.Context) {
	initOnce.Do(func() {
		okCh := make(chan struct{})
		go initDB(ctx, okCh)
		select {
		case <-ctx.Done():
			log.Fatal("Database initialization canceled")
		case <-okCh:
		}
	})
}

func initDB(ctx context.Context, okCh chan struct{}) {
	log.Info("Initializing database...")
	dbType := runtimecfg.Get().Database.Type
	dsn := runtimecfg.Get().Database.DSN

	var db *gorm.DB
	var err error
	gcfg := &gorm.Config{
		Logger:             logger.Default.LogMode(logger.Silent),
		TranslateError:     true,
		PrepareStmt:        true,
		PrepareStmtMaxSize: 2333,
	}
	switch dbType {
	case "sqlite", "sqlite3":
		db, err = gorm.Open(gormlite.Open(dsn), gcfg)
	case "pgsql", "postgres", "postgresql":
		db, err = gorm.Open(postgres.Open(dsn), gcfg)
	case "mysql":
		db, err = gorm.Open(mysql.Open(dsn), gcfg)
	default:
		log.Fatal("unsupported database type", "type", dbType)
	}
	if err != nil {
		log.Fatal("failed to connect database", "err", err)
	}
	err = db.AutoMigrate(
		&entity.Admin{},
		&entity.Artist{},
		&entity.Tag{},
		&entity.TagAlias{},
		&entity.Artwork{},
		&entity.Picture{},
		&entity.UgoiraMeta{},
		&entity.CachedArtwork{},
		&entity.DeletedRecord{},
		&entity.ApiKey{},
		&entity.User{},
	)
	if err != nil {
		log.Fatal("failed to migrate database", "err", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal("failed to get database instance", "err", err)
	}
	if err := sqlDB.PingContext(ctx); err != nil {
		log.Fatal("failed to ping database", "err", err)
	}

	defaultDB = &DB{db: db}
	log.Info("Database initialized")

	okCh <- struct{}{}
}
