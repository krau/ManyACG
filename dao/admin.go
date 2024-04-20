package dao

import (
	"ManyACG-Bot/dao/model"
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var adminCollection *mongo.Collection

func CreateAdmin(ctx context.Context, admin *model.AdminModel) (*mongo.InsertOneResult, error) {
	return adminCollection.InsertOne(ctx, admin)
}

func GetAdminByUserID(ctx context.Context, userID int64) (*model.AdminModel, error) {
	var admin model.AdminModel
	err := adminCollection.FindOne(ctx, bson.M{"user_id": userID}).Decode(&admin)
	if err != nil {
		return nil, err
	}
	return &admin, nil
}

func CreateAdminIfNotExist(ctx context.Context, userID int64) error {
	admin, err := GetAdminByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			_, err := CreateAdmin(ctx, &model.AdminModel{UserID: userID})
			return err
		}
		return err
	}
	if admin != nil {
		return nil
	}
	_, err = CreateAdmin(ctx, &model.AdminModel{UserID: userID})
	return err
}

func DeleteAdminByUserID(ctx context.Context, userID int64) (*mongo.DeleteResult, error) {
	return adminCollection.DeleteOne(ctx, bson.M{"user_id": userID})
}
