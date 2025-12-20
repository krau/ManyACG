package repo

import (
	"context"

	"github.com/krau/ManyACG/internal/model/entity"
	"github.com/krau/ManyACG/internal/shared"
	"github.com/unvgo/ouid"
)

type Admin interface {
	GetAdminByTelegramID(ctx context.Context, tgid int64) (*entity.Admin, error)
	CreateAdmin(ctx context.Context, admin *entity.Admin) (*ouid.OUID, error)
	DeleteAdminByTelegramID(ctx context.Context, tgid int64) error
	ListAdmins(ctx context.Context) ([]entity.Admin, error)
	UpdateAdminPermissions(ctx context.Context, id ouid.OUID, permissions []shared.Permission) error
}
