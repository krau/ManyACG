package service

import (
	"ManyACG/dao"
	"ManyACG/model"
	"ManyACG/types"
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/mongo"
)

func IsAdmin(ctx context.Context, userID int64) (bool, error) {
	admin, err := dao.GetAdminByUserID(ctx, userID)
	if err != nil {
		return false, err
	}
	return admin != nil, nil
}

func CreateAdmin(ctx context.Context, userID int64, permissions []types.Permission, grant int64, super bool) error {
	_, err := dao.CreateAdmin(
		ctx,
		&model.AdminModel{
			UserID:      userID,
			Permissions: permissions,
			GrantBy:     grant,
			SuperAdmin:  super,
		},
	)
	return err
}

func DeleteAdmin(ctx context.Context, userID int64) error {
	_, err := dao.DeleteAdminByUserID(ctx, userID)
	return err
}

func GetAdminByUserID(ctx context.Context, userID int64) (*model.AdminModel, error) {
	return dao.GetAdminByUserID(ctx, userID)
}

func CheckAdminPermission(ctx context.Context, userID int64, permissions ...types.Permission) bool {
	admin, err := dao.GetAdminByUserID(ctx, userID)
	if err != nil {
		return false
	}
	if admin == nil {
		return false
	}
	if admin.SuperAdmin {
		return true
	}
	for _, p := range permissions {
		if !admin.HasPermission(p) {
			return false
		}
	}
	return true
}

func CreateOrUpdateAdmin(ctx context.Context, userID int64, permissions []types.Permission, grant int64, super bool) error {
	admin, err := dao.GetAdminByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return CreateAdmin(ctx, userID, permissions, grant, super)
		}
	}
	if admin == nil {
		return CreateAdmin(ctx, userID, permissions, grant, super)
	}
	admin.Permissions = permissions
	admin.GrantBy = grant
	admin.SuperAdmin = super
	_, err = dao.UpdateAdmin(ctx, admin)
	return err
}
