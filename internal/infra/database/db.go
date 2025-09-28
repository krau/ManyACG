package database

import (
	"context"
	"sync/atomic"

	"github.com/krau/ManyACG/internal/infra/config/runtimecfg"
	"github.com/krau/ManyACG/pkg/log"
	"github.com/ncruces/go-sqlite3/gormlite"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	globalDB atomic.Pointer[gorm.DB]
)

func InitDB(ctx context.Context) {
	log.Info("Initializing database...")
	dbType := runtimecfg.Get().Database.Type
	dsn := runtimecfg.Get().Database.DSN

	var db *gorm.DB
	var err error
	switch dbType {
	case "sqlite":
		db, err = gorm.Open(gormlite.Open(dsn))
	case "pgsql", "postgres", "postgresql":
		db, err = gorm.Open(postgres.Open(dsn))
	case "mysql":
		db, err = gorm.Open(mysql.Open(dsn))
	default:
		log.Fatal("unsupported database type", "type", dbType)
		return
	}
	if err != nil {
		log.Fatal("failed to connect database", "err", err)
		return
	}
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal("failed to get database instance", "err", err)
		return
	}
	if err := sqlDB.PingContext(ctx); err != nil {
		log.Fatal("failed to ping database", "err", err)
		return
	}
	globalDB.Store(db)
	log.Info("Database initialized")
}
