package dao

import (
	"ManyACG/model"
	"ManyACG/types"
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var cachedArtworkCollection *mongo.Collection

func CreateCachedArtwork(ctx context.Context, artwork *types.Artwork, status types.ArtworkStatus) (*mongo.InsertOneResult, error) {
	cachedArtwork := &model.CachedArtworksModel{
		SourceURL: artwork.SourceURL,
		Artwork:   artwork,
		CreatedAt: primitive.NewDateTimeFromTime(time.Now()),
		Status:    status,
	}
	return cachedArtworkCollection.InsertOne(ctx, cachedArtwork)
}

func GetCachedArtworkByURL(ctx context.Context, url string) (*model.CachedArtworksModel, error) {
	var cachedArtwork model.CachedArtworksModel
	err := cachedArtworkCollection.FindOne(ctx, bson.M{"source_url": url}).Decode(&cachedArtwork)
	if err != nil {
		return nil, err
	}
	return &cachedArtwork, err
}

func UpdateCachedArtworkByURL(ctx context.Context, url string, status types.ArtworkStatus) (*mongo.UpdateResult, error) {
	filter := bson.M{"source_url": url}
	update := bson.M{"$set": bson.M{"status": status}}
	return cachedArtworkCollection.UpdateOne(ctx, filter, update)
}

func CleanPostingCachedArtwork(ctx context.Context) (*mongo.DeleteResult, error) {
	filter := bson.M{"status": types.ArtworkStatusPosting}
	return cachedArtworkCollection.DeleteMany(ctx, filter)
}
