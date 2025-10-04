package database

import (
	"context"
	"sync"

	"github.com/krau/ManyACG/internal/infra/config/runtimecfg"
	"github.com/krau/ManyACG/internal/model/entity"
	"github.com/krau/ManyACG/internal/repo"
	"github.com/krau/ManyACG/pkg/log"
	_ "github.com/ncruces/go-sqlite3/embed"
	"github.com/ncruces/go-sqlite3/gormlite"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	defaultDB         *DB
	initOnce          sync.Once
	ErrRecordNotFound = gorm.ErrRecordNotFound
)

type DB struct {
	db *gorm.DB
}

// APIKey implements repo.Repositories.
func (d *DB) APIKey() repo.APIKey {
	return d
}

// Admin implements repo.Repositories.
func (d *DB) Admin() repo.Admin {
	return d
}

// Artist implements repo.Repositories.
func (d *DB) Artist() repo.Artist {
	return d
}

// Artwork implements repo.Repositories.
func (d *DB) Artwork() repo.Artwork {
	return d
}

// CachedArtwork implements repo.Repositories.
func (d *DB) CachedArtwork() repo.CachedArtwork {
	return d
}

// DeletedRecord implements repo.Repositories.
func (d *DB) DeletedRecord() repo.DeletedRecord {
	return d
}

// Picture implements repo.Repositories.
func (d *DB) Picture() repo.Picture {
	return d
}

// Tag implements repo.Repositories.
func (d *DB) Tag() repo.Tag {
	return d
}

// Transaction implements repo.Repositories.
func (d *DB) Transaction(ctx context.Context, fn func(repos repo.Repositories) error) error {
	return d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(&DB{db: tx})
	})
}

func Default() *DB {
	if defaultDB == nil {
		initOnce.Do(func() {
			okCh := make(chan struct{})
			initDB(context.Background(), okCh)
			select {
			case <-context.Background().Done():
				log.Fatal("Database initialization canceled")
			case <-okCh:
			}
		})
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
	switch dbType {
	case "sqlite", "sqlite3":
		db, err = gorm.Open(gormlite.Open(dsn))
	case "pgsql", "postgres", "postgresql":
		db, err = gorm.Open(postgres.Open(dsn))
	case "mysql":
		db, err = gorm.Open(mysql.Open(dsn))
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
