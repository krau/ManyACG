package service

import (
	"context"
	"strings"

	"github.com/krau/ManyACG/internal/infra/storage"
	"github.com/krau/ManyACG/internal/model/dto"
	"github.com/krau/ManyACG/internal/model/entity"
	"github.com/krau/ManyACG/internal/shared"
	"github.com/krau/ManyACG/pkg/strutil"
	"github.com/samber/oops"
)

func (s *Service) Source(sourceType shared.SourceType) storage.Storage {
	return s.storages[shared.StorageType(sourceType)]
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

func (s *Service) FetchArtworkInfo(ctx context.Context, sourceURL string) (*dto.FetchedArtwork, error) {
	for _, sou := range s.sources {
		if _, ok := sou.MatchesSourceURL(sourceURL); ok {
			return sou.GetArtworkInfo(ctx, sourceURL)
		}
	}
	return nil, oops.New("no supported source found")
}

func (s *Service) PrettyFileName(artwork entity.ArtworkLike, picture entity.PictureLike) string {
	for _, sou := range s.sources {
		if _, ok := sou.MatchesSourceURL(artwork.GetSourceURL()); ok {
			return sou.PrettyFileName(artwork, picture)
		}
	}
	ext, _ := strutil.GetFileExtFromURL(picture.GetOriginal())
	return strings.ToLower(strutil.MD5Hash(picture.GetOriginal())) + ext
}
