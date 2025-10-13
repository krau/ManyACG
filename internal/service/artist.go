package service

import (
	"context"

	"github.com/krau/ManyACG/internal/model/entity"
	"github.com/krau/ManyACG/pkg/objectuuid"
)

func (s *Service) GetArtistByID(ctx context.Context, id objectuuid.ObjectUUID) (*entity.Artist, error) {
	return s.repos.Artist().GetArtistByID(ctx, id)
}
