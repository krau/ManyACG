package service

import (
	"context"

	"github.com/krau/ManyACG/internal/model/entity"
	"github.com/krau/ManyACG/internal/shared"
)

func (s *Service) UpdatePictureTelegramInfo(ctx context.Context, pic *entity.Picture, tgInfo *shared.TelegramInfo) error {
	_, err := s.repos.Picture().UpdatePictureTelegramInfoByID(ctx, pic.ID, tgInfo)
	if err != nil {
		return err
	}
	return nil
}
