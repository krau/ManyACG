package dao

import (
	"context"

	"ManyACG/dao/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var artistCollection *mongo.Collection

func CreateArtist(ctx context.Context, artist *model.ArtistModel) (*mongo.InsertOneResult, error) {
	return artistCollection.InsertOne(ctx, artist)
}

func GetArtistByID(ctx context.Context, id primitive.ObjectID) (*model.ArtistModel, error) {
	var artist model.ArtistModel
	if err := artistCollection.FindOne(ctx, bson.M{"_id": id}).Decode(&artist); err != nil {
		return nil, err
	}
	return &artist, nil
}

func GetArtistByUID(ctx context.Context, uid int) (*model.ArtistModel, error) {
	var artist model.ArtistModel
	if err := artistCollection.FindOne(ctx, bson.M{"uid": uid}).Decode(&artist); err != nil {
		return nil, err
	}
	return &artist, nil
}

func GetArtistByName(ctx context.Context, name string) (*model.ArtistModel, error) {
	var artist model.ArtistModel
	if err := artistCollection.FindOne(ctx, bson.M{"name": name}).Decode(&artist); err != nil {
		return nil, err
	}

	return &artist, nil
}

func GetArtistByUserName(ctx context.Context, username string) (*model.ArtistModel, error) {
	var artist model.ArtistModel
	if err := artistCollection.FindOne(ctx, bson.M{"username": username}).Decode(&artist); err != nil {
		return nil, err
	}

	return &artist, nil
}

func GetArtistsByNameLike(ctx context.Context, name string) ([]*model.ArtistModel, error) {
	var artists []*model.ArtistModel
	cursor, err := artistCollection.Find(ctx, bson.M{"name": primitive.Regex{Pattern: name, Options: "i"}})
	if err != nil {
		return nil, err
	}
	if err = cursor.All(ctx, &artists); err != nil {
		return nil, err
	}
	return artists, nil
}

func GetArtistsByUserNameLike(ctx context.Context, username string) ([]*model.ArtistModel, error) {
	var artists []*model.ArtistModel
	cursor, err := artistCollection.Find(ctx, bson.M{"username": primitive.Regex{Pattern: username, Options: "i"}})
	if err != nil {
		return nil, err
	}
	if err = cursor.All(ctx, &artists); err != nil {
		return nil, err
	}
	return artists, nil
}

func UpdateArtistByUID(ctx context.Context, uid int, artist *model.ArtistModel) (*mongo.UpdateResult, error) {
	return artistCollection.UpdateOne(ctx, bson.M{"uid": uid}, bson.M{"$set": artist})
}
