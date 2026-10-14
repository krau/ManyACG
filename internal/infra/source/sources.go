package source

import (
	"context"

	"github.com/krau/ManyACG/internal/model/dto"
	"github.com/krau/ManyACG/internal/shared"
	"github.com/krau/ManyACG/internal/shared/errs"
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

type LocalFetcher map[shared.SourceType]ArtworkSource

func (sources LocalFetcher) GetArtworkInfo(ctx context.Context, sourceURL string) (*dto.FetchedArtwork, error) {
	for _, source := range sources {
		if canonicalURL, ok := source.MatchesSourceURL(sourceURL); ok {
			return source.GetArtworkInfo(ctx, canonicalURL)
		}
	}
	return nil, errs.ErrUnsupportedSource
}
