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

func (s *Service) FindSourceURL(text string) string {
	text = strings.ReplaceAll(text, "\n", " ")
	for _, sou := range s.namers {
		if url, ok := sou.MatchesSourceURL(text); ok {
			return url
		}
	}
	return ""
}

// GetArtworkInfo fetches artwork details from the configured fetcher.
func (s *Service) GetArtworkInfo(ctx context.Context, sourceURL string) (*dto.FetchedArtwork, error) {
	if s.fetcher == nil {
		return nil, errs.ErrUnsupportedSource
	}
	return s.fetcher.GetArtworkInfo(ctx, sourceURL)
}

func (s *Service) PrettyFileName(artwork shared.ArtworkLike, picture shared.PictureLike) string {
	for _, sou := range s.namers {
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
