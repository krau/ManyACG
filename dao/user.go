package dao

import (
	"context"
	"time"

	"github.com/krau/ManyACG/types"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var (
	userCollection *mongo.Collection
)

func CreateUser(ctx context.Context, user *types.UserModel) (*mongo.InsertOneResult, error) {
	user.UpdatedAt = primitive.NewDateTimeFromTime(time.Now())
	if user.Settings == nil {
		user.Settings = &types.UserSettings{}
	}
	return userCollection.InsertOne(ctx, user)
}

func GetUserByID(ctx context.Context, id primitive.ObjectID) (*types.UserModel, error) {
	user := &types.UserModel{}
	err := userCollection.FindOne(ctx, bson.M{"_id": id}).Decode(user)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func GetUserByUsername(ctx context.Context, username string) (*types.UserModel, error) {
	user := &types.UserModel{}
	err := userCollection.FindOne(ctx, bson.M{"username": bson.M{"$regex": primitive.Regex{Pattern: "^" + username + "$", Options: "i"}}}).Decode(user)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func GetUserByTelegramID(ctx context.Context, telegramID int64) (*types.UserModel, error) {
	user := &types.UserModel{}
	err := userCollection.FindOne(ctx, bson.M{"telegram_id": telegramID}).Decode(user)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func GetUserByEmail(ctx context.Context, email string) (*types.UserModel, error) {
	user := &types.UserModel{}
	err := userCollection.FindOne(ctx, bson.M{"email": bson.M{"$regex": primitive.Regex{Pattern: "^" + email + "$", Options: "i"}}}).Decode(user)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func UpdateUserSettings(ctx context.Context, id primitive.ObjectID, settings *types.UserSettings) (*mongo.UpdateResult, error) {
	// if settings == nil {
	// 	return nil, manyacgErrors.ErrSettingsNil
	// }

	// updateDoc := bson.M{}
	// v := reflect.ValueOf(settings).Elem()
	// t := v.Type()

	// for i := 0; i < v.NumField(); i++ {
	// 	field := v.Field(i)
	// 	fieldName := t.Field(i).Tag.Get("bson")

	// 	if field.Kind() == reflect.Bool {
	// 		updateDoc["settings."+fieldName] = field.Interface()
	// 	} else if !field.IsZero() {
	// 		updateDoc["settings."+fieldName] = field.Interface()
	// 	}
	// }

	// result, err := userCollection.UpdateOne(ctx, bson.M{"_id": id}, bson.M{
	// 	"$set": updateDoc,
	// })
	// if err != nil {
	// 	return nil, err
	// }

	// return result, nil

	return userCollection.UpdateOne(ctx, bson.M{"_id": id}, bson.M{
		"$set": bson.M{"settings": settings},
	})
}
