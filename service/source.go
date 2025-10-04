package service

import (
	"strings"

	"github.com/krau/ManyACG/internal/infra/storage"
	"github.com/krau/ManyACG/internal/shared"
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