package dao

import (
	"context"
	"time"

	"github.com/krau/ManyACG/model"
	"github.com/krau/ManyACG/types"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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

func UpdateCachedArtworkStatusByURL(ctx context.Context, url string, status types.ArtworkStatus) (*mongo.UpdateResult, error) {
	filter := bson.M{"source_url": url}
	update := bson.M{"$set": bson.M{"status": status}}
	return cachedArtworkCollection.UpdateOne(ctx, filter, update)
}

func CleanPostingCachedArtwork(ctx context.Context) (*mongo.DeleteResult, error) {
	filter := bson.M{"status": types.ArtworkStatusPosting}
	return cachedArtworkCollection.DeleteMany(ctx, filter)
}

func UpdateCachedArtwork(ctx context.Context, artwork *model.CachedArtworksModel) (*mongo.UpdateResult, error) {
	filter := bson.M{"source_url": artwork.SourceURL}
	update := bson.M{"$set": bson.M{"artwork": artwork.Artwork}}
	artwork.CreatedAt = primitive.NewDateTimeFromTime(time.Now())
	return cachedArtworkCollection.UpdateOne(ctx, filter, update, options.Update().SetUpsert(true))
}

func DeleteCachedArtworkByURL(ctx context.Context, url string) (*mongo.DeleteResult, error) {
	filter := bson.M{"source_url": url}
	return cachedArtworkCollection.DeleteOne(ctx, filter)
}
