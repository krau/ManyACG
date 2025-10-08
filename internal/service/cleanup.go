package service

import (
	"context"
)

func (s *Service) Cleanup(ctx context.Context) error {
	return s.repos.CachedArtwork().ResetPostingCachedArtworkStatus(ctx)
}
