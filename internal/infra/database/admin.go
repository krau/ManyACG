package database

import (
	"context"

	"github.com/krau/ManyACG/internal/model/entity"
	"github.com/krau/ManyACG/internal/shared"
	"github.com/krau/ManyACG/pkg/objectuuid"
	"gorm.io/gorm"
)

func (d *DB) GetAdminByTelegramID(ctx context.Context, telegramID int64) (*entity.Admin, error) {
	res, err := gorm.G[entity.Admin](d.db).Where("telegram_id = ?", telegramID).First(ctx)
	if err != nil {
		return nil, err
	}
	return &res, err
}

func (d *DB) CreateAdmin(ctx context.Context, admin *entity.Admin) (*objectuuid.ObjectUUID, error) {
	result := gorm.WithResult()
	err := gorm.G[entity.Admin](d.db, result).Create(ctx, admin)
	if err != nil {
		return nil, err
	}
	return &admin.ID, nil
}

func (d *DB) DeleteAdminByTelegramID(ctx context.Context, telegramID int64) error {
	n, err := gorm.G[entity.Admin](d.db).Where("telegram_id = ?", telegramID).Delete(ctx)
	if err != nil {
		return err
	}
	if n == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (d *DB) ListAdmins(ctx context.Context) ([]entity.Admin, error) {
	admins, err := gorm.G[entity.Admin](d.db).Find(ctx)
	if err != nil {
		return nil, err
	}
	return admins, nil
}

func (d *DB) UpdateAdminPermissions(ctx context.Context, id objectuuid.ObjectUUID, permissions []shared.Permission) error {
	return d.db.WithContext(ctx).Model(&entity.Admin{}).Where("id = ?", id).Update("permissions", permissions).Error
}
