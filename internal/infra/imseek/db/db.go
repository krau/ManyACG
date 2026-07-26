package db

import (
	"context"
	"database/sql"
	"fmt"
	"path/filepath"

	_ "github.com/ncruces/go-sqlite3/driver"
)

const schema = `
CREATE TABLE IF NOT EXISTS image (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    hash       BLOB UNIQUE NOT NULL,
    path       TEXT NOT NULL,
    artwork_id TEXT NOT NULL DEFAULT '',
    phash      TEXT NOT NULL DEFAULT '',
    dhash      TEXT NOT NULL DEFAULT ''
);
CREATE TABLE IF NOT EXISTS vector_stats (
    id                 INTEGER PRIMARY KEY,
    vector_count       INTEGER NOT NULL,
    total_vector_count INTEGER NOT NULL,
    indexed            BOOLEAN NOT NULL DEFAULT 0
);
CREATE TABLE IF NOT EXISTS vector (
    id     INTEGER PRIMARY KEY,
    vector BLOB NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_vector_stats_total_vector_count ON vector_stats (total_vector_count);
CREATE INDEX IF NOT EXISTS idx_vector_stats_indexed ON vector_stats (indexed);
CREATE INDEX IF NOT EXISTS idx_image_path ON image (path);
CREATE INDEX IF NOT EXISTS idx_image_phash_dhash ON image (phash, dhash);
CREATE INDEX IF NOT EXISTS idx_image_artwork_id ON image (artwork_id);
`

type DB struct {
	sql *sql.DB
}

func Open(dir string, wal bool) (*DB, error) {
	path := filepath.Join(dir, "imseek.db")
	journal := "WAL"
	if !wal {
		journal = "DELETE"
	}
	dsn := fmt.Sprintf("file:%s?_pragma=journal_mode(%s)&_pragma=synchronous(NORMAL)&_pragma=busy_timeout(5000)",
		path, journal)
	sqlDB, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, err
	}
	if _, err := sqlDB.ExecContext(context.Background(), schema); err != nil {
		sqlDB.Close()
		return nil, fmt.Errorf("init schema: %w", err)
	}
	return &DB{sql: sqlDB}, nil
}

func (d *DB) Close() error { return d.sql.Close() }

func (d *DB) SQL() *sql.DB { return d.sql }
