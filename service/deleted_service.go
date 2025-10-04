package service

import "context"

func (s *Service) CheckDeletedByURL(ctx context.Context, sourceURL string) bool {
	return s.repos.DeletedRecord().CheckDeletedByURL(ctx, sourceURL)
}

func (s *Service) CancelDeletedByURL(ctx context.Context, sourceURL string) error {
	return s.repos.DeletedRecord().DeleteDeletedByURL(ctx, sourceURL)
}
