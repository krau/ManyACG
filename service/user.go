package service

import (
	"ManyACG/dao"
	"ManyACG/model"
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func GetUserByID(ctx context.Context, id primitive.ObjectID) (*model.UserModel, error) {
	return dao.GetUserByID(ctx, id)
}

func GetUserByUsername(ctx context.Context, username string) (*model.UserModel, error) {
	return dao.GetUserByUsername(ctx, username)
}

func CreateUser(ctx context.Context, user *model.UserModel) (*model.UserModel, error) {
	res, err := dao.CreateUser(ctx, user)
	if err != nil {
		return nil, err
	}
	return dao.GetUserByID(ctx, res.InsertedID.(primitive.ObjectID))
}
