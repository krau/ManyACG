package dao

import (
	"ManyACG/model"
	"ManyACG/types"
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func QueryArtworksByTexts(ctx context.Context, texts [][]string, r18 types.R18Type, limit int64) ([]*model.ArtworkModel, error) {
	if len(texts) == 0 {
		return GetArtworksByR18(ctx, r18, limit)
	}
	var artworks []*model.ArtworkModel
	var cursor *mongo.Cursor
	var err error
	var andConditions []bson.M

	for _, textGroup := range texts {
		var orConditions []bson.M
		tagIDs := getTagIDs(ctx, textGroup)
		if len(tagIDs) > 0 {
			orConditions = append(orConditions, bson.M{"tags": bson.M{"$in": tagIDs}})
		}

		artistIDs := getArtistIDs(ctx, textGroup)
		if len(artistIDs) > 0 {
			orConditions = append(orConditions, bson.M{"artist_id": bson.M{"$in": artistIDs}})
		}

		for _, text := range textGroup {
			orConditions = append(orConditions, bson.M{"title": bson.M{"$regex": text, "$options": "i"}})
			orConditions = append(orConditions, bson.M{"description": bson.M{"$regex": text, "$options": "i"}})
		}

		if len(orConditions) > 0 {
			andConditions = append(andConditions, bson.M{"$or": orConditions})
		}
	}

	match := bson.M{"$and": andConditions}
	if r18 == types.R18TypeAll {
		cursor, err = artworkCollection.Aggregate(ctx, mongo.Pipeline{
			bson.D{{Key: "$match", Value: match}},
			bson.D{{Key: "$sample", Value: bson.M{"size": limit}}},
		})
	} else {
		match["r18"] = r18 == types.R18TypeOnly
		cursor, err = artworkCollection.Aggregate(ctx, mongo.Pipeline{
			bson.D{{Key: "$match", Value: match}},
			bson.D{{Key: "$sample", Value: bson.M{"size": limit}}},
		})
	}
	if err != nil {
		return nil, err
	}
	err = cursor.All(ctx, &artworks)
	if err != nil {
		return nil, err
	}
	if len(artworks) == 0 {
		return nil, mongo.ErrNoDocuments
	}
	return artworks, nil
}

func getTagIDs(ctx context.Context, tags []string) []primitive.ObjectID {
	var tagIDs []primitive.ObjectID
	for _, tag := range tags {
		tagModels, err := QueryTagsByName(ctx, tag)
		if err == nil {
			for _, tagModel := range tagModels {
				tagIDs = append(tagIDs, tagModel.ID)
			}
		}
	}
	return tagIDs
}

func getArtistIDs(ctx context.Context, artists []string) []primitive.ObjectID {
	var artistIDs []primitive.ObjectID
	for _, artist := range artists {
		artistModels, err := QueryArtistsByName(ctx, artist)
		if err == nil {
			for _, artistModel := range artistModels {
				artistIDs = append(artistIDs, artistModel.ID)
			}
		}
		artistModelsByUsername, err := QueryArtistsByUserName(ctx, artist)
		if err == nil {
			for _, artistModel := range artistModelsByUsername {
				artistIDs = append(artistIDs, artistModel.ID)
			}
		}
	}
	return artistIDs
}
