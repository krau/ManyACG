// Admin 中的 Create 全部是如果不存在则创建. 如果存在则返回 nil

package dao

import (
	"ManyACG/dao/model"
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var adminCollection *mongo.Collection

func GetAdminByUserID(ctx context.Context, userID int64) (*model.AdminModel, error) {
	var admin model.AdminModel
	err := adminCollection.FindOne(ctx, bson.M{"user_id": userID}).Decode(&admin)
	if err != nil {
		return nil, err
	}
	return &admin, nil
}

func CreateAdmin(ctx context.Context, admin *model.AdminModel) (*mongo.InsertOneResult, error) {
	_, err := GetAdminByUserID(ctx, admin.UserID)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return adminCollection.InsertOne(ctx, admin)
		}
		return nil, err
	}
	return nil, nil
}

func CreateSuperAdminByUserID(ctx context.Context, userID int64, grant int64) (*mongo.InsertOneResult, error) {
	return CreateAdmin(ctx, &model.AdminModel{UserID: userID, GrantBy: grant, SuperAdmin: true})
}

func DeleteAdminByUserID(ctx context.Context, userID int64) (*mongo.DeleteResult, error) {
	return adminCollection.DeleteOne(ctx, bson.M{"user_id": userID})
}

func UpdateAdmin(ctx context.Context, admin *model.AdminModel) (*mongo.UpdateResult, error) {
	return adminCollection.ReplaceOne(ctx, bson.M{"user_id": admin.UserID}, admin)
}
