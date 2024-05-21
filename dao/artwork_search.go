package dao

import (
	"ManyACG/model"
	"ManyACG/types"
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// like 查询
func QueryArtworksByTitle(ctx context.Context, title string, r18 types.R18Type, limit int64) ([]*model.ArtworkModel, error) {
	if title == "" {
		return GetArtworksByR18(ctx, r18, limit)
	}
	var artworks []*model.ArtworkModel
	var cursor *mongo.Cursor
	var err error
	switch r18 {
	case types.R18TypeAll:
		cursor, err = artworkCollection.Aggregate(ctx, mongo.Pipeline{
			bson.D{{Key: "$match", Value: bson.M{"title": primitive.Regex{Pattern: title, Options: "i"}}}},
			bson.D{{Key: "$sample", Value: bson.M{"size": limit}}},
		})
	case types.R18TypeOnly:
		cursor, err = artworkCollection.Aggregate(ctx, mongo.Pipeline{
			bson.D{{Key: "$match", Value: bson.M{"title": primitive.Regex{Pattern: title, Options: "i"}, "r18": true}}},
			bson.D{{Key: "$sample", Value: bson.M{"size": limit}}},
		})
	case types.R18TypeNone:
		cursor, err = artworkCollection.Aggregate(ctx, mongo.Pipeline{
			bson.D{{Key: "$match", Value: bson.M{"title": primitive.Regex{Pattern: title, Options: "i"}, "r18": false}}},
			bson.D{{Key: "$sample", Value: bson.M{"size": limit}}},
		})
	}
	if err != nil {
		return nil, err
	}
	if err = cursor.All(ctx, &artworks); err != nil {
		return nil, err
	}
	if len(artworks) == 0 {
		return nil, mongo.ErrNoDocuments
	}
	return artworks, nil
}

func QueryArtworksByDescription(ctx context.Context, description string, r18 types.R18Type, limit int64) ([]*model.ArtworkModel, error) {
	if description == "" {
		return GetArtworksByR18(ctx, r18, limit)
	}
	var artworks []*model.ArtworkModel
	var cursor *mongo.Cursor
	var err error
	switch r18 {
	case types.R18TypeAll:
		cursor, err = artworkCollection.Aggregate(ctx, mongo.Pipeline{
			bson.D{{Key: "$match", Value: bson.M{"description": primitive.Regex{Pattern: description, Options: "i"}}}},
			bson.D{{Key: "$sample", Value: bson.M{"size": limit}}},
		})
	case types.R18TypeOnly:
		cursor, err = artworkCollection.Aggregate(ctx, mongo.Pipeline{
			bson.D{{Key: "$match", Value: bson.M{"description": primitive.Regex{Pattern: description, Options: "i"}, "r18": true}}},
			bson.D{{Key: "$sample", Value: bson.M{"size": limit}}},
		})
	case types.R18TypeNone:
		cursor, err = artworkCollection.Aggregate(ctx, mongo.Pipeline{
			bson.D{{Key: "$match", Value: bson.M{"description": primitive.Regex{Pattern: description, Options: "i"}, "r18": false}}},
			bson.D{{Key: "$sample", Value: bson.M{"size": limit}}},
		})
	}
	if err != nil {
		return nil, err
	}
	if err = cursor.All(ctx, &artworks); err != nil {
		return nil, err
	}
	if len(artworks) == 0 {
		return nil, mongo.ErrNoDocuments
	}
	return artworks, nil
}
