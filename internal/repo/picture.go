package repo

import (
	"context"

	"github.com/krau/ManyACG/internal/model/entity"
	"github.com/krau/ManyACG/internal/model/query"
	"github.com/krau/ManyACG/internal/shared"
	"github.com/unvgo/ouid"
)

type Picture interface {
	GetPictureByID(ctx context.Context, id ouid.OUID) (*entity.Picture, error)
	DeletePictureByID(ctx context.Context, id ouid.OUID) error
	UpdatePictureTelegramInfoByID(ctx context.Context, id ouid.OUID, tgInfo *shared.TelegramInfo) (*entity.Picture, error)
	QueryPicturesByPhash(ctx context.Context, que query.PicturesPhash) ([]*entity.Picture, error)
	RandomPictures(ctx context.Context, limit int) ([]*entity.Picture, error)
	SavePicture(ctx context.Context, pic *entity.Picture) error
}
