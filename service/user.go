package service

import (
	"ManyACG-Bot/dao"
	"context"
)

func IsAdmin(ctx context.Context, userID int64) (bool, error) {
	admin, err := dao.GetAdminByUserID(ctx, userID)
	if err != nil {
		return false, err
	}
	return admin != nil, nil
}

func CreateAdmin(ctx context.Context, userID int64) error {
	return dao.CreateAdminIfNotExist(ctx, userID)
}

func DeleteAdmin(ctx context.Context, userID int64) error {
	_, err := dao.DeleteAdminByUserID(ctx, userID)
	return err
}

func SetAdmin(ctx context.Context, userID int64) error {
	admin, err := dao.GetAdminByUserID(ctx, userID)
	if err != nil {
		if admin == nil {
			return CreateAdmin(ctx, userID)
		}
		return err
	}
	if admin != nil {
		_, err := dao.DeleteAdminByUserID(ctx, userID)
		return err
	}
	return nil
}
