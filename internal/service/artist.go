package service

import (
	"context"

	"github.com/krau/ManyACG/internal/model/entity"
	"github.com/unvgo/ouid"
)

func (s *Service) GetArtistByID(ctx context.Context, id ouid.OUID) (*entity.Artist, error) {
	return s.repos.Artist().GetArtistByID(ctx, id)
}
