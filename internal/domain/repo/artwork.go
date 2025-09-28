package repo

import (
	"context"

	"github.com/krau/ManyACG/internal/domain/entity/artist"
	"github.com/krau/ManyACG/internal/domain/entity/artwork"
	"github.com/krau/ManyACG/internal/domain/entity/tag"
	"gorm.io/gorm"
)

type ArtworkRepo interface {
	Save(ctx context.Context, artwork *artwork.Artwork) error
	FindByURL(ctx context.Context, sourceURL string) (*artwork.Artwork, error)
}

type ArtistRepo interface {
	Save(ctx context.Context, artist *artist.Artist) error
	FindBySourceAndUID(ctx context.Context, sourceType string, uid string) (*artist.Artist, error)
}

type TagRepo interface {
	Save(ctx context.Context, tag *tag.Tag) error
	FindByNameWithAlias(ctx context.Context, find string) (*tag.Tag, error)
}

type TransactionRepo interface {
	WithTransaction(ctx context.Context, fn func(repos *TransactionRepos) error) error
}

type TransactionRepos struct {
	ArtworkRepo ArtworkRepo
	ArtistRepo  ArtistRepo
	TagRepo     TagRepo
}

var ErrNotFound = gorm.ErrRecordNotFound
