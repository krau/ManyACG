package repo

import (
	"context"

	"github.com/krau/ManyACG/internal/model/entity"
	"github.com/krau/ManyACG/internal/shared"
	"github.com/krau/ManyACG/pkg/objectuuid"
)

type Ugoira interface {
	UpdateUgoiraTelegramInfoByID(ctx context.Context, id objectuuid.ObjectUUID, tgInfo *shared.TelegramInfo) (*entity.UgoiraMeta, error)
}
