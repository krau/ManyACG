package service

import (
	"context"
	"errors"

	"github.com/krau/ManyACG/internal/model/entity"
	"github.com/krau/ManyACG/internal/shared"
	"github.com/krau/ManyACG/internal/shared/errs"
)

func (s *Service) IsAdminByTgID(ctx context.Context, tgid int64) (bool, error) {
	admin, err := s.repos.Admin().GetAdminByTelegramID(ctx, tgid)
	if err != nil {
		if errors.Is(err, errs.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	return admin != nil, nil
}

func (s *Service) CreateAdmin(ctx context.Context, tgID int64, permissions []shared.Permission) error {
	exist, err := s.repos.Admin().GetAdminByTelegramID(ctx, tgID)
	if err != nil && !errors.Is(err, errs.ErrRecordNotFound) {
		return err
	}
	if exist != nil {
		return nil
	}
	_, err = s.repos.Admin().CreateAdmin(ctx, &entity.Admin{
		TelegramID:  tgID,
		Permissions: permissions,
	})
	return err
}

func (s *Service) CheckAdminPermissionByTgID(ctx context.Context, userID int64, permissions ...shared.Permission) bool {
	admin, err := s.repos.Admin().GetAdminByTelegramID(ctx, userID)
	if err != nil {
		return false
	}
	if admin == nil {
		return false
	}
	isSudo := false
	for _, p := range admin.Permissions {
		if p == shared.PermissionSudo {
			isSudo = true
			break
		}
	}
	if isSudo {
		return true
	}
	for _, p := range permissions {
		has := false
		for _, ap := range admin.Permissions {
			if ap == p {
				has = true
				break
			}
		}
		if !has {
			return false
		}
	}
	return true
}

func (s *Service) CreateOrUpdateAdmin(ctx context.Context, tgid int64, permissions []shared.Permission) error {
	exist, err := s.repos.Admin().GetAdminByTelegramID(ctx, tgid)
	if err != nil && !errors.Is(err, errs.ErrRecordNotFound) {
		return err
	}
	if exist == nil {
		_, err = s.repos.Admin().CreateAdmin(ctx, &entity.Admin{
			TelegramID:  tgid,
			Permissions: permissions,
		})
		return err
	}
	err = s.repos.Admin().UpdateAdminPermissions(ctx, exist.ID, permissions)
	return err
}

func (s *Service) GetAdminUserIDs(ctx context.Context) ([]int64, error) {
	admins, err := s.repos.Admin().ListAdmins(ctx)
	if err != nil {
		return nil, err
	}
	var userIDs []int64
	for _, admin := range admins {
		if admin.TelegramID > 0 {
			userIDs = append(userIDs, admin.TelegramID)
		}
	}
	return userIDs, nil
}

func (s *Service) GetAdminGroupIDs(ctx context.Context) ([]int64, error) {
	admins, err := s.repos.Admin().ListAdmins(ctx)
	if err != nil {
		return nil, err
	}
	var groupIDs []int64
	for _, admin := range admins {
		if admin.TelegramID < 0 {
			groupIDs = append(groupIDs, admin.TelegramID)
		}
	}
	return groupIDs, nil
}

func (s *Service) GetAdminByTelegramID(ctx context.Context, tgid int64) (*entity.Admin, error) {
	return s.repos.Admin().GetAdminByTelegramID(ctx, tgid)
}

func (s *Service) DeleteAdminByTgID(ctx context.Context, tgid int64) error {
	return s.repos.Admin().DeleteAdminByTelegramID(ctx, tgid)
}
