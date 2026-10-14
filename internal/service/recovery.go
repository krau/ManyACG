package service

import (
	"context"
)

func (s *Service) RecoverPosting(ctx context.Context) error {
	return s.repos.CachedArtwork().ResetPostingCachedArtworkStatus(ctx)
}
