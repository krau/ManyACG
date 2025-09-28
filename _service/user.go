package service

import (
	"context"

	"github.com/krau/ManyACG/dao"
	"github.com/krau/ManyACG/types"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func GetUserByID(ctx context.Context, id primitive.ObjectID) (*types.UserModel, error) {
	return dao.GetUserByID(ctx, id)
}

func GetUserByUsername(ctx context.Context, username string) (*types.UserModel, error) {
	return dao.GetUserByUsername(ctx, username)
}

func GetUserByTelegramID(ctx context.Context, telegramID int64) (*types.UserModel, error) {
	return dao.GetUserByTelegramID(ctx, telegramID)
}

func GetUserByEmail(ctx context.Context, email string) (*types.UserModel, error) {
	return dao.GetUserByEmail(ctx, email)
}

func CreateUser(ctx context.Context, user *types.UserModel) (*types.UserModel, error) {
	res, err := dao.CreateUser(ctx, user)
	if err != nil {
		return nil, err
	}
	return dao.GetUserByID(ctx, res.InsertedID.(primitive.ObjectID))
}

func CreateUnauthUser(ctx context.Context, user *types.UnauthUserModel) (*types.UnauthUserModel, error) {
	res, err := dao.CreateUnauthUser(ctx, user)
	if err != nil {
		return nil, err
	}
	return dao.GetUnauthUserByID(ctx, res.InsertedID.(primitive.ObjectID))
}

func GetUnauthUserByID(ctx context.Context, id primitive.ObjectID) (*types.UnauthUserModel, error) {
	return dao.GetUnauthUserByID(ctx, id)
}

func GetUnauthUserByUsername(ctx context.Context, username string) (*types.UnauthUserModel, error) {
	return dao.GetUnauthUserByUsername(ctx, username)
}

func UpdateUnauthUser(ctx context.Context, id primitive.ObjectID, user *types.UnauthUserModel) (*types.UnauthUserModel, error) {
	_, err := dao.UpdateUnauthUser(ctx, id, user)
	if err != nil {
		return nil, err
	}
	return dao.GetUnauthUserByID(ctx, id)
}

func DeleteUnauthUser(ctx context.Context, id primitive.ObjectID) error {
	_, err := dao.DeleteUnauthUser(ctx, id)
	return err
}

func UpdateUserSettings(ctx context.Context, id primitive.ObjectID, settings *types.UserSettings) (*types.UserSettings, error) {
	_, err := dao.UpdateUserSettings(ctx, id, settings)
	if err != nil {
		return nil, err
	}
	user, err := dao.GetUserByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return user.Settings, nil
}
