package dao

import (
	"ManyACG-Bot/dao/collections"
	"ManyACG-Bot/types"
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var artworkCollection *mongo.Collection = DB.Collection(collections.Artworks)

func CreateArtwork(ctx context.Context, artwork *types.Artwork) (*mongo.InsertOneResult, error) {
	return artworkCollection.InsertOne(ctx, artwork)
}

func GetArtworkByURL(ctx context.Context, url string) (*types.Artwork, error) {
	var artwork types.Artwork
	err := artworkCollection.FindOne(ctx, bson.M{"source.url": url}).Decode(&artwork)
	return &artwork, err
}
