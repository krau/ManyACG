package repo

import (
	"context"

	"github.com/krau/ManyACG/internal/model/entity"
	"github.com/krau/ManyACG/internal/shared"
	"github.com/krau/ManyACG/pkg/objectuuid"
)

type Admin interface {
	GetAdminByTelegramID(ctx context.Context, tgid int64) (*entity.Admin, error)
	CreateAdmin(ctx context.Context, admin *entity.Admin) (*objectuuid.ObjectUUID, error)
	DeleteAdminByTelegramID(ctx context.Context, tgid int64) error
	ListAdmins(ctx context.Context) ([]entity.Admin, error)
	UpdateAdminPermissions(ctx context.Context, id objectuuid.ObjectUUID, permissions []shared.Permission) error
}
