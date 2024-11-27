package types

import (
	"context"
)

type Service interface {
	GetArtworkByURL(ctx context.Context, url string, opts ...*AdapterOption) (*Artwork, error)
}
