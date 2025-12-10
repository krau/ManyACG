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
	if err := optimizePgSQL(ctx, db); err != nil {
		log.Fatal("failed to optimize pgsql database", "err", err)
	}

	defaultDB = &DB{db: db}
	log.Info("Database initialized")

	okCh <- struct{}{}
}

func optimizePgSQL(ctx context.Context, db *gorm.DB) error {
	if db.Dialector.Name() != "postgres" {
		return nil
	}
	pgroongaEnabled := runtimecfg.Get().Database.Pgsql.PGroonga
	if pgroongaEnabled {
		// use pgroonga indexes
		log.Debug("Applying pgroonga index optimizations...")
		sqls := []string{
			`CREATE EXTENSION IF NOT EXISTS pgroonga;`,

			// delete existing trigram indexes if any
			`DROP INDEX IF EXISTS idx_artworks_title_trgm;`,
			`DROP INDEX IF EXISTS idx_artworks_description_trgm;`,
			`DROP INDEX IF EXISTS idx_artists_name_trgm;`,
			`DROP INDEX IF EXISTS idx_tags_name_trgm;`,
			`DROP INDEX IF EXISTS idx_tag_alias_alias_trgm;`,
			// PGroonga indexes
			`CREATE INDEX IF NOT EXISTS pgroonga_artworks_title ON artworks USING pgroonga (title);`,
			`CREATE INDEX IF NOT EXISTS pgroonga_artworks_description ON artworks USING pgroonga (description);`,
			`CREATE INDEX IF NOT EXISTS pgroonga_artists_name ON artists USING pgroonga (name);`,
			`CREATE INDEX IF NOT EXISTS pgroonga_tags_name ON tags USING pgroonga (name);`,
			`CREATE INDEX IF NOT EXISTS pgroonga_tag_alias_alias ON tag_aliases USING pgroonga (alias);`,

			// Foreign key
			`CREATE INDEX IF NOT EXISTS idx_artworks_artist_id ON artworks(artist_id);`,
			`CREATE INDEX IF NOT EXISTS idx_artwork_tags_artwork_id ON artwork_tags(artwork_id);`,
			`CREATE INDEX IF NOT EXISTS idx_artwork_tags_tag_id ON artwork_tags(tag_id);`,
		}
		for _, sql := range sqls {
			if err := db.WithContext(ctx).Exec(sql).Error; err != nil {
				return err
			}
		}
		log.Debug("pgroonga index optimizations applied")
		if runtimecfg.Get().Database.Pgsql.ReindexPGroonga {
			log.Debug("Reindexing pgroonga indexes...")
			pgroongaIndexes := []string{
				"pgroonga_artworks_title",
				"pgroonga_artworks_description",
				"pgroonga_artists_name",
				"pgroonga_tags_name",
				"pgroonga_tag_alias_alias",
			}
			for _, index := range pgroongaIndexes {
				reindexSQL := "REINDEX INDEX " + index + ";"
				if err := db.WithContext(ctx).Exec(reindexSQL).Error; err != nil {
					return err
				}
			}
			log.Debug("Reindexing of pgroonga indexes completed")
		}
		return nil
	}
	// by default, use pg_trgm indexes
	log.Debug("Applying pg_trgm index optimizations...")

	sqls := []string{
		// Extensions
		`CREATE EXTENSION IF NOT EXISTS pg_trgm;`,
		// delete existing pgroonga indexes if any
		`DROP INDEX IF EXISTS pgroonga_artworks_title;`,
		`DROP INDEX IF EXISTS pgroonga_artworks_description;`,
		`DROP INDEX IF EXISTS pgroonga_artists_name;`,
		`DROP INDEX IF EXISTS pgroonga_tags_name;`,
		`DROP INDEX IF EXISTS pgroonga_tag_alias_alias;`,

		// Trigram indexes
		`CREATE INDEX IF NOT EXISTS idx_artworks_title_trgm ON artworks USING gin (title gin_trgm_ops);`,
		`CREATE INDEX IF NOT EXISTS idx_artworks_description_trgm ON artworks USING gin (description gin_trgm_ops);`,
		`CREATE INDEX IF NOT EXISTS idx_artists_name_trgm ON artists USING gin (name gin_trgm_ops);`,
		`CREATE INDEX IF NOT EXISTS idx_tags_name_trgm ON tags USING gin (name gin_trgm_ops);`,
		`CREATE INDEX IF NOT EXISTS idx_tag_alias_alias_trgm ON tag_aliases USING gin (alias gin_trgm_ops);`,

		// Foreign key
		`CREATE INDEX IF NOT EXISTS idx_artworks_artist_id ON artworks(artist_id);`,
		`CREATE INDEX IF NOT EXISTS idx_artwork_tags_artwork_id ON artwork_tags(artwork_id);`,
		`CREATE INDEX IF NOT EXISTS idx_artwork_tags_tag_id ON artwork_tags(tag_id);`,
	}
	for _, sql := range sqls {
		if err := db.WithContext(ctx).Exec(sql).Error; err != nil {
			return err
		}
	}
	log.Debug("pg_trgm index optimizations applied")
	return nil
}
