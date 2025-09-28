package database

import (
	"context"

	"github.com/krau/ManyACG/internal/domain/entity/admin"
	"github.com/krau/ManyACG/internal/domain/repo"
	"github.com/krau/ManyACG/pkg/objectuuid"
	"gorm.io/gorm"
)

type adminRepo struct {
	db *gorm.DB
}

// Delete implements repo.AdminRepo.
func (a *adminRepo) Delete(ctx context.Context, admin *admin.Admin) error {
	panic("unimplemented")
}

// FindByID implements repo.AdminRepo.
func (a *adminRepo) FindByID(ctx context.Context, id objectuuid.ObjectUUID) (*admin.Admin, error) {
	panic("unimplemented")
}

// FindByTelegramID implements repo.AdminRepo.
func (a *adminRepo) FindByTelegramID(ctx context.Context, telegramID int64) (*admin.Admin, error) {
	panic("unimplemented")
}

// IsAdmin implements repo.AdminRepo.
func (a *adminRepo) IsAdmin(ctx context.Context, userID int64) (bool, error) {
	panic("unimplemented")
}

// Save implements repo.AdminRepo.
func (a *adminRepo) Save(ctx context.Context, admin *admin.Admin) error {
	panic("unimplemented")
}

func NewAdminRepo(db *gorm.DB) *adminRepo {
	return &adminRepo{db: db}
}

var _ repo.AdminRepo = (*adminRepo)(nil)
