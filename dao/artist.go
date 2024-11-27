package dao

import (
	"context"

	"github.com/krau/ManyACG/types"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var artistCollection *mongo.Collection

func CreateArtist(ctx context.Context, artist *types.ArtistModel) (*mongo.InsertOneResult, error) {
	return artistCollection.InsertOne(ctx, artist)
}

func GetArtistByID(ctx context.Context, id primitive.ObjectID) (*types.ArtistModel, error) {
	var artist types.ArtistModel
	if err := artistCollection.FindOne(ctx, bson.M{"_id": id}).Decode(&artist); err != nil {
		return nil, err
	}
	return &artist, nil
}

func GetArtistByUserName(ctx context.Context, username string, sourceType types.SourceType) (*types.ArtistModel, error) {
	var artist types.ArtistModel
	if err := artistCollection.FindOne(ctx, bson.M{"username": username, "type": string(sourceType)}).Decode(&artist); err != nil {
		return nil, err
	}
	return &artist, nil
}

func GetArtistCount(ctx context.Context) (int64, error) {
	return artistCollection.CountDocuments(ctx, bson.M{})
}

func QueryArtistsByName(ctx context.Context, name string) ([]*types.ArtistModel, error) {
	if name == "" {
		return nil, mongo.ErrNoDocuments
	}
	var artists []*types.ArtistModel
	cursor, err := artistCollection.Find(ctx, bson.M{"name": primitive.Regex{Pattern: name, Options: "i"}})
	if err != nil {
		return nil, err
	}
	if err = cursor.All(ctx, &artists); err != nil {
		return nil, err
	}
	if len(artists) == 0 {
		return nil, mongo.ErrNoDocuments
	}
	return artists, nil
}

func QueryArtistsByUserName(ctx context.Context, username string) ([]*types.ArtistModel, error) {
	if username == "" {
		return nil, mongo.ErrNoDocuments
	}
	var artists []*types.ArtistModel
	cursor, err := artistCollection.Find(ctx, bson.M{"username": primitive.Regex{Pattern: username, Options: "i"}})
	if err != nil {
		return nil, err
	}
	if err = cursor.All(ctx, &artists); err != nil {
		return nil, err
	}
	if len(artists) == 0 {
		return nil, mongo.ErrNoDocuments
	}
	return artists, nil
}

func GetArtistByUID(ctx context.Context, uid string, sourceType types.SourceType) (*types.ArtistModel, error) {
	var artist types.ArtistModel
	if err := artistCollection.FindOne(ctx, bson.M{"uid": uid, "type": string(sourceType)}).Decode(&artist); err != nil {
		return nil, err
	}
	return &artist, nil
}

func UpdateArtist(ctx context.Context, artist *types.ArtistModel) (*mongo.UpdateResult, error) {
	return artistCollection.UpdateOne(ctx, bson.M{"_id": artist.ID}, bson.M{"$set": artist})
}
