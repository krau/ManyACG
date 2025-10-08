package service

import (
	"context"

	"github.com/krau/ManyACG/internal/model/entity"
)

func (s *Service) CheckDeletedByURL(ctx context.Context, sourceURL string) bool {
	return s.repos.DeletedRecord().CheckDeletedByURL(ctx, sourceURL)
}

func (s *Service) CancelDeletedByURL(ctx context.Context, sourceURL string) error {
	return s.repos.DeletedRecord().DeleteDeletedByURL(ctx, sourceURL)
}


func (s *Service) GetDeletedByURL(ctx context.Context, sourceURL string) (*entity.DeletedRecord, error) {
	return s.repos.DeletedRecord().GetDeletedByURL(ctx, sourceURL)
}