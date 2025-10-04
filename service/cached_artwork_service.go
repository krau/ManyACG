package service

import "context"

func (s *Service) DeleteCachedArtworkByURL(ctx context.Context, sourceURL string) error {
	cachedArt, err := s.repos.CachedArtwork().GetCachedArtworkByURL(ctx, sourceURL)
	if err != nil {
		return err
	}
	return s.repos.CachedArtwork().DeleteCachedArtworkByID(ctx, cachedArt.ID)
}
