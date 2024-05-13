package dao

import (
	"ManyACG/dao/model"
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

func GetRandomArtworks(ctx context.Context, limit int) ([]*model.ArtworkModel, error) {
	var artworks []*model.ArtworkModel
	cursor, err := artworkCollection.Aggregate(ctx, mongo.Pipeline{
		bson.D{{Key: "$sample", Value: bson.M{"size": limit}}},
	})
	if err != nil {
		return nil, err
	}
	err = cursor.All(ctx, &artworks)
	if err != nil {
		return nil, err
	}
	return artworks, nil

}

func GetRandomArtworksR18(ctx context.Context, r18 bool, limit int) ([]*model.ArtworkModel, error) {
	var artworks []*model.ArtworkModel
	cursor, err := artworkCollection.Aggregate(ctx, mongo.Pipeline{
		bson.D{{Key: "$match", Value: bson.M{"r18": r18}}},
		bson.D{{Key: "$sample", Value: bson.M{"size": limit}}},
	})
	if err != nil {
		return nil, err
	}
	err = cursor.All(ctx, &artworks)
	if err != nil {
		return nil, err
	}
	return artworks, nil
}

func GetRandomArtworksByTags(ctx context.Context, tags []primitive.ObjectID, limit int) ([]*model.ArtworkModel, error) {
	var artworks []*model.ArtworkModel
	cursor, err := artworkCollection.Aggregate(ctx, mongo.Pipeline{
		bson.D{{Key: "$match", Value: bson.M{"tags": bson.M{"$all": tags}}}},
		bson.D{{Key: "$sample", Value: bson.M{"size": limit}}},
	})
	if err != nil {
		return nil, err
	}
	err = cursor.All(ctx, &artworks)
	if err != nil {
		return nil, err
	}
	return artworks, nil
}

func GetRandomArtworksByTagsR18(ctx context.Context, tags []primitive.ObjectID, r18 bool, limit int) ([]*model.ArtworkModel, error) {
	var artworks []*model.ArtworkModel
	cursor, err := artworkCollection.Aggregate(ctx, mongo.Pipeline{
		bson.D{{Key: "$match", Value: bson.M{"tags": bson.M{"$all": tags}, "r18": r18}}},
		bson.D{{Key: "$sample", Value: bson.M{"size": limit}}},
	})
	if err != nil {
		return nil, err
	}
	err = cursor.All(ctx, &artworks)
	if err != nil {
		return nil, err
	}
	return artworks, nil
}

func UpdateArtworkPicturesByID(ctx context.Context, id primitive.ObjectID, pictures []primitive.ObjectID) (*mongo.UpdateResult, error) {
	return artworkCollection.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": bson.M{"pictures": pictures}})
}

func DeleteArtworkByID(ctx context.Context, id primitive.ObjectID) (*mongo.DeleteResult, error) {
	return artworkCollection.DeleteOne(ctx, bson.M{"_id": id})
}

func DeleteArtworkPicturesByID(ctx context.Context, artworkID primitive.ObjectID, pictureIDs []primitive.ObjectID) (*mongo.UpdateResult, error) {
	return artworkCollection.UpdateOne(ctx, bson.M{"_id": artworkID}, bson.M{"$pull": bson.M{"pictures": bson.M{"$in": pictureIDs}}})
}
