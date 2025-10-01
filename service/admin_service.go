package service

import (
	"context"
	"errors"

	"github.com/krau/ManyACG/internal/model/entity"
	"github.com/krau/ManyACG/internal/shared"
	"github.com/krau/ManyACG/internal/shared/errs"
)

func (s *Service) IsAdminByTgID(ctx context.Context, tgid int64) (bool, error) {
	// admin, err := dao.GetAdminByUserID(ctx, userID)
	// if err != nil {
	// 	return false, err
	// }
	// return admin != nil, nil
	admin, err := s.repos.Admin().GetAdminByTelegramID(ctx, tgid)
	if err != nil {
		if errors.Is(err, errs.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	return admin != nil, nil
}

func (s *Service) CreateAdmin(ctx context.Context, userID int64, permissions []shared.Permission, grant int64, super bool) error {
	// _, err := dao.CreateAdmin(
	// 	ctx, &types.AdminModel{
	// 		UserID:      userID,
	// 		Permissions: permissions,
	// 		GrantBy:     grant,
	// 		SuperAdmin:  super,
	// 	},
	// )
	// return err
	exist, err := s.repos.Admin().GetAdminByTelegramID(ctx, userID)
	if err != nil && !errors.Is(err, errs.ErrRecordNotFound) {
		return err
	}
	if exist != nil {
		return nil
	}
	_, err = s.repos.Admin().CreateAdmin(ctx, &entity.Admin{
		TelegramID:  userID,
		Permissions: permissions,
	})
	return err
}

// func DeleteAdminByTgID(ctx context.Context, tgid int64) error {
// 	// _, err := dao.DeleteAdminByUserID(ctx, userID)
// 	// return err
// 	return database.Default().DeleteAdminByTelegramID(ctx, tgid)
// }

// func GetAdminByTgID(ctx context.Context, userID int64) (*entity.Admin, error) {
// 	// return dao.GetAdminByUserID(ctx, userID)
// 	admin, err := database.Default().GetAdminByTelegramID(ctx, userID)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return admin, nil
// }

func (s *Service) CheckAdminPermissionByTgID(ctx context.Context, userID int64, permissions ...shared.Permission) bool {
	// admin, err := dao.GetAdminByUserID(ctx, userID)
	// if err != nil {
	// 	return false
	// }
	// if admin == nil {
	// 	return false
	// }
	// if admin.SuperAdmin {
	// 	return true
	// }
	// for _, p := range permissions {
	// 	if !admin.HasPermission(p) {
	// 		return false
	// 	}
	// }
	// return true
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

func (s *Service) CreateOrUpdateAdmin(ctx context.Context, tgid int64, permissions []shared.Permission, grant int64, super bool) error {
	// admin, err := dao.GetAdminByUserID(ctx, userID)
	// if err != nil {
	// 	if errors.Is(err, mongo.ErrNoDocuments) {
	// 		return CreateAdmin(ctx, userID, permissions, grant, super)
	// 	}
	// }
	// if admin == nil {
	// 	return CreateAdmin(ctx, userID, permissions, grant, super)
	// }
	// admin.Permissions = permissions
	// admin.GrantBy = grant
	// admin.SuperAdmin = super
	// _, err = dao.UpdateAdmin(ctx, admin)
	// return err
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
	// admins, err := dao.GetAdmins(ctx)
	// if err != nil {
	// 	return nil, err
	// }
	// var userIDs []int64
	// for _, admin := range admins {
	// 	if admin.UserID > 0 {
	// 		userIDs = append(userIDs, admin.UserID)
	// 	}
	// }
	// return userIDs, nil
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
	// admins, err := dao.GetAdmins(ctx)
	// if err != nil {
	// 	return nil, err
	// }
	// var groupIDs []int64
	// for _, admin := range admins {
	// 	if admin.UserID < 0 {
	// 		groupIDs = append(groupIDs, admin.UserID)
	// 	}
	// }
	// return groupIDs, nil
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
