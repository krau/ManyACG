package dao

import (
	"ManyACG/model"
	"ManyACG/types"
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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

func GetArtworksByR18(ctx context.Context, r18 types.R18Type, limit int) ([]*model.ArtworkModel, error) {
	var artworks []*model.ArtworkModel
	var cursor *mongo.Cursor
	var err error
	if r18 == types.R18TypeAll {
		cursor, err = artworkCollection.Aggregate(ctx, mongo.Pipeline{
			bson.D{{Key: "$sample", Value: bson.M{"size": limit}}},
		})
	} else {
		cursor, err = artworkCollection.Aggregate(ctx, mongo.Pipeline{
			bson.D{{Key: "$match", Value: bson.M{"r18": r18 == types.R18TypeOnly}}},
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

func GetArtworksByTags(ctx context.Context, tags [][]primitive.ObjectID, r18 types.R18Type, limit int) ([]*model.ArtworkModel, error) {
	if len(tags) == 0 {
		return GetArtworksByR18(ctx, r18, limit)
	}
	var artworks []*model.ArtworkModel
	var cursor *mongo.Cursor
	var err error
	match := bson.M{}
	var orConditions []bson.M
	for _, tagGroup := range tags {
		var orCondition []bson.M
		for _, tag := range tagGroup {
			orCondition = append(orCondition, bson.M{"tags": tag})
		}
		orConditions = append(orConditions, bson.M{"$or": orCondition})
	}
	match["$and"] = orConditions
	if r18 == types.R18TypeAll {
		cursor, err = artworkCollection.Aggregate(ctx, mongo.Pipeline{
			bson.D{{Key: "$match", Value: match}},
			bson.D{{Key: "$sample", Value: bson.M{"size": limit}}},
		})
	} else {
		cursor, err = artworkCollection.Aggregate(ctx, mongo.Pipeline{
			bson.D{{Key: "$match", Value: bson.M{"r18": r18 == types.R18TypeOnly}}},
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

func GetArtworksByArtistID(ctx context.Context, artistID primitive.ObjectID, r18 types.R18Type, page, pageSize int64) ([]*model.ArtworkModel, error) {
	var artworks []*model.ArtworkModel
	var cursor *mongo.Cursor
	var err error
	opts := options.Find().SetSort(bson.M{"_id": -1}).SetSkip((page - 1) * pageSize).SetLimit(pageSize)
	if r18 == types.R18TypeAll {
		cursor, err = artworkCollection.Find(ctx, bson.M{"artist_id": artistID}, opts)
	} else {
		cursor, err = artworkCollection.Find(ctx, bson.M{"artist_id": artistID, "r18": r18 == types.R18TypeOnly}, opts)
	}
	if err != nil {
		return nil, err
	}
	err = cursor.All(ctx, &artworks)
	if err != nil {
		return nil, err
	}
	return artworks, nil
}

func GetArtworkCount(ctx context.Context, r18 types.R18Type) (int64, error) {
	if r18 == types.R18TypeAll {
		return artworkCollection.CountDocuments(ctx, bson.M{})
	}
	return artworkCollection.CountDocuments(ctx, bson.M{"r18": r18 == types.R18TypeOnly})
}

func GetLatestArtworks(ctx context.Context, r18 types.R18Type, page, pageSize int64) ([]*model.ArtworkModel, error) {
	var artworks []*model.ArtworkModel
	var cursor *mongo.Cursor
	var err error
	opts := options.Find().SetSort(bson.M{"_id": -1}).SetSkip((page - 1) * pageSize).SetLimit(pageSize)
	if r18 == types.R18TypeAll {
		cursor, err = artworkCollection.Find(ctx, bson.M{}, opts)
	} else {
		cursor, err = artworkCollection.Find(ctx, bson.M{"r18": r18 == types.R18TypeOnly}, opts)
	}
	if err != nil {
		return nil, err
	}
	err = cursor.All(ctx, &artworks)
	if err != nil {
		return nil, err
	}
	return artworks, nil
}

func GetArtworkCountByArtistID(ctx context.Context, artistID primitive.ObjectID, r18 types.R18Type) (int64, error) {
	if r18 == types.R18TypeAll {
		return artworkCollection.CountDocuments(ctx, bson.M{"artist_id": artistID})
	}
	return artworkCollection.CountDocuments(ctx, bson.M{"artist_id": artistID, "r18": r18 == types.R18TypeOnly})
}

func UpdateArtworkPicturesByID(ctx context.Context, id primitive.ObjectID, pictures []primitive.ObjectID) (*mongo.UpdateResult, error) {
	return artworkCollection.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": bson.M{"pictures": pictures}})
}

func UpdateArtworkR18ByID(ctx context.Context, id primitive.ObjectID, r18 bool) (*mongo.UpdateResult, error) {
	return artworkCollection.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": bson.M{"r18": r18}})
}

func UpdateArtworkTagsByID(ctx context.Context, id primitive.ObjectID, tags []primitive.ObjectID) (*mongo.UpdateResult, error) {
	return artworkCollection.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": bson.M{"tags": tags}})
}

func UpdateArtworkTitleByID(ctx context.Context, id primitive.ObjectID, title string) (*mongo.UpdateResult, error) {
	return artworkCollection.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": bson.M{"title": title}})
}

func IncrementArtworkLikeCountByID(ctx context.Context, id primitive.ObjectID) (*mongo.UpdateResult, error) {
	return artworkCollection.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$inc": bson.M{"like_count": int64(1)}})
}

func DeleteArtworkByID(ctx context.Context, id primitive.ObjectID) (*mongo.DeleteResult, error) {
	return artworkCollection.DeleteOne(ctx, bson.M{"_id": id})
}
