package domain

import (
	"context"
)

type ArtworkRepo interface {
	Save(ctx context.Context, artwork *Artwork) error
}
