package repo

import (
	"context"

	"github.com/krau/ManyACG/internal/domain/entity/admin"
	"github.com/krau/ManyACG/pkg/objectuuid"
)

type AdminRepo interface {
	Save(ctx context.Context, admin *admin.Admin) error
	FindByTelegramID(ctx context.Context, telegramID int64) (*admin.Admin, error)
	FindByID(ctx context.Context, id objectuuid.ObjectUUID) (*admin.Admin, error)
	IsAdmin(ctx context.Context, tgUserID int64) (bool, error)
	Delete(ctx context.Context, admin *admin.Admin) error
}
