package repo

import (
	"context"

	"github.com/krau/ManyACG/internal/model/entity"
	"github.com/krau/ManyACG/internal/shared"
	"github.com/krau/ManyACG/pkg/objectuuid"
)

type Picture interface {
	GetPictureByID(ctx context.Context, id objectuuid.ObjectUUID) (*entity.Picture, error)
	DeletePictureByID(ctx context.Context, id objectuuid.ObjectUUID) error
	UpdatePictureTelegramInfoByID(ctx context.Context, id objectuuid.ObjectUUID, tgInfo *shared.TelegramInfo) (*entity.Picture, error)
}
