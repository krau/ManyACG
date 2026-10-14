package service

import (
	"context"
	"maps"
	"strings"

	"github.com/krau/ManyACG/internal/infra/source"
	"github.com/krau/ManyACG/internal/model/dto"
	"github.com/krau/ManyACG/internal/shared"
	"github.com/krau/ManyACG/internal/shared/errs"
	"github.com/krau/ManyACG/pkg/strutil"
)

func (s *Service) Source(sourceType shared.SourceType) source.ArtworkSource {
	return s.sources[sourceType]
}

func (s *Service) FindSourceURL(text string) string {
	text = strings.ReplaceAll(text, "\n", " ")
	for _, sou := range s.sources {
		if url, ok := sou.MatchesSourceURL(text); ok {
			return url
		}
	}
	return ""
}

// GetArtworkInfo fetches artwork details from the configured remote fetcher,
// or from the matching local source when no remote fetcher is configured.
func (s *Service) GetArtworkInfo(ctx context.Context, sourceURL string) (*dto.FetchedArtwork, error) {
	if s.fetcher != nil {
		return s.fetcher.GetArtworkInfo(ctx, sourceURL)
	}
	for _, sou := range s.sources {
		if canonicalURL, ok := sou.MatchesSourceURL(sourceURL); ok {
			return sou.GetArtworkInfo(ctx, canonicalURL)
		}
	}
	return nil, errs.ErrUnsupportedSource
}

func (s *Service) PrettyFileName(artwork shared.ArtworkLike, picture shared.PictureLike) string {
	for _, sou := range s.sources {
		if _, ok := sou.MatchesSourceURL(artwork.GetSourceURL()); ok {
			return sou.PrettyFileName(artwork, picture)
		}
	}
	ext, _ := strutil.GetFileExtFromURL(picture.GetOriginal())
	return strings.ToLower(strutil.MD5Hash(picture.GetOriginal())) + ext
}

func (s *Service) SourcePollers() map[shared.SourceType]source.ArtworkPoller {
	return maps.Clone(s.pollers)
}
