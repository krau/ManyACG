package source

import (
	"context"

	"github.com/krau/ManyACG/internal/model/dto"
	"github.com/krau/ManyACG/internal/shared"
)

type ArtworkNamer interface {
	MatchesSourceURL(sourceUrl string) (string, bool)
	PrettyFileName(artwork shared.ArtworkLike, picture shared.PictureLike) string
}

type ArtworkSource interface {
	ArtworkNamer
	GetArtworkInfo(ctx context.Context, sourceUrl string) (*dto.FetchedArtwork, error)
	FetchNewArtworks(ctx context.Context, limit int) ([]*dto.FetchedArtwork, error)
}
