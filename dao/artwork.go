package dao

import (
	"ManyACG-Bot/types"
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var artworkCollection *mongo.Collection

func CreateArtwork(ctx context.Context, artwork *types.Artwork) (*mongo.InsertOneResult, error) {
	artwork.CreatedAt = time.Now()
	return artworkCollection.InsertOne(ctx, artwork)
}

func GetArtworkByURL(ctx context.Context, url string) (*types.Artwork, error) {
	var artwork types.Artwork
	err := artworkCollection.FindOne(ctx, bson.M{"source.url": url}).Decode(&artwork)
	return &artwork, err
}

func GetArtworkByPicture(ctx context.Context, picture *types.Picture) (*types.Artwork, error) {
	var artwork types.Artwork
	err := artworkCollection.FindOne(ctx, bson.M{"pictures.original": picture.Original}).Decode(&artwork)
	return &artwork, err
}
