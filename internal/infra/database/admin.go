package database

import (
	"context"

	"github.com/krau/ManyACG/internal/infra/database/model"
	"github.com/krau/ManyACG/internal/shared"
	"github.com/krau/ManyACG/pkg/objectuuid"
	"gorm.io/gorm"
)

func (d *DB) GetAdminByTelegramID(ctx context.Context, telegramID int64) (*model.Admin, error) {
	res, err := gorm.G[model.Admin](d.db).Where("telegram_id = ?", telegramID).First(ctx)
	return &res, err
}

func (d *DB) CreateAdmin(ctx context.Context, admin *model.Admin) (*objectuuid.ObjectUUID, error) {
	result := gorm.WithResult()
	err := gorm.G[model.Admin](d.db, result).Create(ctx, admin)
	if err != nil {
		return nil, err
	}
	return &admin.ID, nil
}

func (d *DB) DeleteAdminByTelegramID(ctx context.Context, telegramID int64) error {
	n, err := gorm.G[model.Admin](d.db).Where("telegram_id = ?", telegramID).Delete(ctx)
	if err != nil {
		return err
	}
	if n == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (d *DB) ListAdmins(ctx context.Context) ([]model.Admin, error) {
	admins, err := gorm.G[model.Admin](d.db).Find(ctx)
	if err != nil {
		return nil, err
	}
	return admins, nil
}

func (d *DB) UpdateAdminPermissions(ctx context.Context, id objectuuid.ObjectUUID, permissions []shared.Permission) error {
	return d.db.WithContext(ctx).Model(&model.Admin{}).Where("id = ?", id).Update("permissions", permissions).Error
}
