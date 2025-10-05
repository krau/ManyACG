package service

import (
	"context"

	"github.com/krau/ManyACG/internal/model/entity"
	"github.com/krau/ManyACG/internal/model/query"
	"github.com/krau/ManyACG/internal/shared"
	"github.com/krau/ManyACG/pkg/objectuuid"
)

func (s *Service) UpdatePictureTelegramInfo(ctx context.Context, id objectuuid.ObjectUUID, tgInfo *shared.TelegramInfo) error {
	_, err := s.repos.Picture().UpdatePictureTelegramInfoByID(ctx, id, tgInfo)
	if err != nil {
		return err
	}
	return nil
}

func (s *Service) QueryPicturesByPhash(ctx context.Context, que query.PicturesPhash) ([]*entity.Picture, error) {
	return s.repos.Picture().QueryPicturesByPhash(ctx, que)
}
