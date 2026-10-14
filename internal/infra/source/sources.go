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

type ArtworkFetcher interface {
	GetArtworkInfo(ctx context.Context, sourceUrl string) (*dto.FetchedArtwork, error)
}

type ArtworkSource interface {
	ArtworkNamer
	ArtworkFetcher
}

type ArtworkPoller interface {
	FetchNewArtworks(ctx context.Context, limit int) ([]*dto.FetchedArtwork, error)
}
