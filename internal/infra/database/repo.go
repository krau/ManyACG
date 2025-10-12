package database

import (
	"context"

	"github.com/krau/ManyACG/internal/repo"
	"gorm.io/gorm"
)

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

func (d *DB) Ugoira() repo.Ugoira {
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
