package dao

import (
	"ManyACG/dao/model"
	"ManyACG/types"
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var cachedArtworkCollection *mongo.Collection

func CreateCachedArtwork(ctx context.Context, artwork *types.Artwork) (*mongo.InsertOneResult, error) {
	cachedArtwork := &model.CachedArtworksModel{
		SourceURL: artwork.SourceURL,
		Artwork:   artwork,
		CreatedAt: primitive.NewDateTimeFromTime(time.Now()),
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
