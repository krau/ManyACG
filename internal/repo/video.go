package repo

import (
	"context"

	"github.com/krau/ManyACG/internal/model/entity"
	"github.com/krau/ManyACG/internal/shared"
	"github.com/unvgo/ouid"
)

type Video interface {
	UpdateVideoTelegramInfoByID(ctx context.Context, id ouid.OUID, tgInfo *shared.TelegramInfo) (*entity.Video, error)
}
