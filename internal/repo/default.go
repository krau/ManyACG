package repo

import (
	"context"
)

type Repositories interface {
	Admin() Admin
	APIKey() APIKey
	Artist() Artist
	Artwork() Artwork
	Tag() Tag
	Picture() Picture
	Ugoira() Ugoira
	DeletedRecord() DeletedRecord
	CachedArtwork() CachedArtwork
	Transactional
}

type Transactional interface {
	Transaction(ctx context.Context, fn func(repos Repositories) error) error
}
