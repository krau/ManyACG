package dao

import (
	"ManyACG-Bot/dao/model"
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var artworkCollection *mongo.Collection

func CreateArtwork(ctx context.Context, artwork *model.ArtworkModel) (*mongo.InsertOneResult, error) {
	artwork.CreatedAt = primitive.NewDateTimeFromTime(time.Now())
	return artworkCollection.InsertOne(ctx, artwork)
}

func GetArtworkByID(ctx context.Context, id primitive.ObjectID) (*model.ArtworkModel, error) {
	var artwork model.ArtworkModel
	err := artworkCollection.FindOne(ctx, bson.M{"_id": id}).Decode(&artwork)
	if err != nil {
		return nil, err
	}
	return &artwork, err
}

func GetArtworkByURL(ctx context.Context, url string) (*model.ArtworkModel, error) {
	var artwork model.ArtworkModel
	err := artworkCollection.FindOne(ctx, bson.M{"source_url": url}).Decode(&artwork)
	if err != nil {
		return nil, err
	}
	return &artwork, err
}

func UpdateArtworkPicturesByID(ctx context.Context, id primitive.ObjectID, pictures []primitive.ObjectID) (*mongo.UpdateResult, error) {
	return artworkCollection.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": bson.M{"pictures": pictures}})
}

func DeleteArtworkByID(ctx context.Context, id primitive.ObjectID) (*mongo.DeleteResult, error) {
	return artworkCollection.DeleteOne(ctx, bson.M{"_id": id})
}
