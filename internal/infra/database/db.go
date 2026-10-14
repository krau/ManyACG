package database

import (
	"context"
	"fmt"
	"sync"

	"github.com/krau/ManyACG/internal/infra/config/runtimecfg"
	"github.com/krau/ManyACG/internal/model/entity"
	"github.com/krau/ManyACG/pkg/log"
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

func (d *DB) Close() error {
	sqlDB, err := d.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

func Init(ctx context.Context, cfg runtimecfg.Config) {
	initOnce.Do(func() {
		okCh := make(chan struct{})
		go initDB(ctx, cfg, okCh)
		select {
		case <-ctx.Done():
			log.Fatal("Database initialization canceled")
		case <-okCh:
		}
	})
}

func initDB(ctx context.Context, cfg runtimecfg.Config, okCh chan struct{}) {
	log.Info("Initializing database...")
	db, err := openDatabase(ctx, cfg)
	if err != nil {
		log.Fatal("failed to initialize database", "err", err)
	}
	if !cfg.App.Mode.ReadOnly() && cfg.Database.AutoMigrate {
		if err := migrateDatabase(ctx, db, cfg); err != nil {
			log.Fatal("failed to migrate database", "err", err)
		}
	}
	defaultDB = &DB{db: db}
	log.Info("Database initialized")
	okCh <- struct{}{}
}

func openDatabase(ctx context.Context, cfg runtimecfg.Config) (*gorm.DB, error) {
	dbType := cfg.Database.Type
	dsn := cfg.Database.DSN

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
		return nil, fmt.Errorf("unsupported database type %q", dbType)
	}
	if err != nil {
		return nil, fmt.Errorf("connect database: %w", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("get database instance: %w", err)
	}
	if err := sqlDB.PingContext(ctx); err != nil {
		_ = sqlDB.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	return db, nil
}

func Migrate(ctx context.Context, cfg runtimecfg.Config) error {
	db, err := openDatabase(ctx, cfg)
	if err != nil {
		return err
	}
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	defer sqlDB.Close()
	return migrateDatabase(ctx, db, cfg)
}

func migrateDatabase(ctx context.Context, db *gorm.DB, cfg runtimecfg.Config) error {
	if err := db.WithContext(ctx).AutoMigrate(
		&entity.Admin{},
		&entity.Artist{},
		&entity.Tag{},
		&entity.TagAlias{},
		&entity.Artwork{},
		&entity.Picture{},
		&entity.UgoiraMeta{},
		&entity.Video{},
		&entity.CachedArtwork{},
		&entity.DeletedRecord{},
		&entity.ApiKey{},
		&entity.User{},
	); err != nil {
		return fmt.Errorf("migrate database: %w", err)
	}
	return optimizePgSQL(ctx, db, cfg)
}

func optimizePgSQL(ctx context.Context, db *gorm.DB, cfg runtimecfg.Config) error {
	if db.Dialector.Name() != "postgres" {
		return nil
	}
	pgroongaEnabled := cfg.Database.Pgsql.PGroonga
	if pgroongaEnabled {
		// use pgroonga indexes
		log.Debug("Applying pgroonga index optimizations...")
		sqls := []string{
			`CREATE EXTENSION IF NOT EXISTS pgroonga;`,

			// PGroonga indexes
			`CREATE INDEX IF NOT EXISTS pgroonga_artworks_title_like ON artworks USING pgroonga (title pgroonga_text_regexp_ops_v2);`,
			`CREATE INDEX IF NOT EXISTS pgroonga_artworks_description_like ON artworks USING pgroonga (description pgroonga_text_regexp_ops_v2);`,
			`CREATE INDEX IF NOT EXISTS pgroonga_artists_name_like ON artists USING pgroonga (name pgroonga_text_regexp_ops_v2);`,
			`CREATE INDEX IF NOT EXISTS pgroonga_tags_name_like ON tags USING pgroonga (name pgroonga_text_regexp_ops_v2);`,
			`CREATE INDEX IF NOT EXISTS pgroonga_tag_alias_alias_like ON tag_aliases USING pgroonga (alias pgroonga_text_regexp_ops_v2);`,

			// Foreign key
			`CREATE INDEX IF NOT EXISTS idx_artworks_artist_id ON artworks(artist_id);`,
			`CREATE INDEX IF NOT EXISTS idx_artwork_tags_artwork_id ON artwork_tags(artwork_id);`,
			`CREATE INDEX IF NOT EXISTS idx_artwork_tags_tag_artwork ON artwork_tags(tag_id,artwork_id);`,

			// Delete obsolete indexes only after their replacements exist.
			`DROP INDEX IF EXISTS pgroonga_artworks_title;`,
			`DROP INDEX IF EXISTS pgroonga_artworks_description;`,
			`DROP INDEX IF EXISTS pgroonga_artists_name;`,
			`DROP INDEX IF EXISTS pgroonga_tags_name;`,
			`DROP INDEX IF EXISTS pgroonga_tag_alias_alias;`,
			`DROP INDEX IF EXISTS idx_artworks_title_trgm;`,
			`DROP INDEX IF EXISTS idx_artworks_description_trgm;`,
			`DROP INDEX IF EXISTS idx_artists_name_trgm;`,
			`DROP INDEX IF EXISTS idx_tags_name_trgm;`,
			`DROP INDEX IF EXISTS idx_tag_alias_alias_trgm;`,
			`DROP INDEX IF EXISTS idx_artwork_tags_tag_id;`,
		}
		for _, sql := range sqls {
			if err := db.WithContext(ctx).Exec(sql).Error; err != nil {
				return err
			}
		}
		log.Debug("pgroonga index optimizations applied")
		if cfg.Database.Pgsql.ReindexPGroonga {
			log.Debug("Reindexing pgroonga indexes...")
			pgroongaIndexes := []string{
				"pgroonga_artworks_title_like",
				"pgroonga_artworks_description_like",
				"pgroonga_artists_name_like",
				"pgroonga_tags_name_like",
				"pgroonga_tag_alias_alias_like",
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
		// Trigram indexes
		`CREATE INDEX IF NOT EXISTS idx_artworks_title_trgm ON artworks USING gin (title gin_trgm_ops);`,
		`CREATE INDEX IF NOT EXISTS idx_artworks_description_trgm ON artworks USING gin (description gin_trgm_ops);`,
		`CREATE INDEX IF NOT EXISTS idx_artists_name_trgm ON artists USING gin (name gin_trgm_ops);`,
		`CREATE INDEX IF NOT EXISTS idx_tags_name_trgm ON tags USING gin (name gin_trgm_ops);`,
		`CREATE INDEX IF NOT EXISTS idx_tag_alias_alias_trgm ON tag_aliases USING gin (alias gin_trgm_ops);`,

		// Foreign key
		`CREATE INDEX IF NOT EXISTS idx_artworks_artist_id ON artworks(artist_id);`,
		`CREATE INDEX IF NOT EXISTS idx_artwork_tags_artwork_id ON artwork_tags(artwork_id);`,
		`CREATE INDEX IF NOT EXISTS idx_artwork_tags_tag_artwork ON artwork_tags(tag_id,artwork_id);`,

		// Delete obsolete indexes only after their replacements exist.
		`DROP INDEX IF EXISTS pgroonga_artworks_title;`,
		`DROP INDEX IF EXISTS pgroonga_artworks_description;`,
		`DROP INDEX IF EXISTS pgroonga_artists_name;`,
		`DROP INDEX IF EXISTS pgroonga_tags_name;`,
		`DROP INDEX IF EXISTS pgroonga_tag_alias_alias;`,
		`DROP INDEX IF EXISTS pgroonga_artworks_title_like;`,
		`DROP INDEX IF EXISTS pgroonga_artworks_description_like;`,
		`DROP INDEX IF EXISTS pgroonga_artists_name_like;`,
		`DROP INDEX IF EXISTS pgroonga_tags_name_like;`,
		`DROP INDEX IF EXISTS pgroonga_tag_alias_alias_like;`,
		`DROP INDEX IF EXISTS idx_artwork_tags_tag_id;`,
	}
	for _, sql := range sqls {
		if err := db.WithContext(ctx).Exec(sql).Error; err != nil {
			return err
		}
	}
	log.Debug("pg_trgm index optimizations applied")
	return nil
}
