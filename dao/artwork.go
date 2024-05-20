package dao

import (
	"ManyACG/dao/model"
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

func GetArtworksR18(ctx context.Context, r18 bool, limit int) ([]*model.ArtworkModel, error) {
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

func GetArtworksByTags(ctx context.Context, tags []primitive.ObjectID, limit int) ([]*model.ArtworkModel, error) {
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

func GetArtworksByTagsR18(ctx context.Context, tags []primitive.ObjectID, r18 bool, limit int) ([]*model.ArtworkModel, error) {
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

func GetArtworksByTitle(ctx context.Context, title string, limit int) ([]*model.ArtworkModel, error) {
	var artworks []*model.ArtworkModel
	cursor, err := artworkCollection.Find(ctx, bson.M{"title": primitive.Regex{Pattern: title, Options: "i"}}, options.Find().SetLimit(int64(limit)))
	if err != nil {
		return nil, err
	}
	err = cursor.All(ctx, &artworks)
	if err != nil {
		return nil, err
	}
	return artworks, nil
}

func GetArtworksByDescription(ctx context.Context, description string, limit int) ([]*model.ArtworkModel, error) {
	var artworks []*model.ArtworkModel
	cursor, err := artworkCollection.Find(ctx, bson.M{"description": primitive.Regex{Pattern: description, Options: "i"}}, options.Find().SetLimit(int64(limit)))
	if err != nil {
		return nil, err
	}
	err = cursor.All(ctx, &artworks)
	if err != nil {
		return nil, err
	}
	return artworks, nil
}

func GetArtworksByArtistID(ctx context.Context, artistID primitive.ObjectID, limit int) ([]*model.ArtworkModel, error) {
	var artworks []*model.ArtworkModel
	cursor, err := artworkCollection.Find(ctx, bson.M{"artist._id": artistID}, options.Find().SetLimit(int64(limit)))
	if err != nil {
		return nil, err
	}
	err = cursor.All(ctx, &artworks)
	if err != nil {
		return nil, err
	}
	return artworks, nil
}

func GetArtworksByArtistName(ctx context.Context, artistName string, limit int) ([]*model.ArtworkModel, error) {
	var artworks []*model.ArtworkModel
	artists, err := GetArtistsByNameLike(ctx, artistName)
	if err != nil {
		return nil, err
	}
	artistIDs := make([]primitive.ObjectID, len(artists))
	for i, artist := range artists {
		artistIDs[i] = artist.ID
	}
	cursor, err := artworkCollection.Find(ctx, bson.M{"artist_id": bson.M{"$in": artistIDs}}, options.Find().SetLimit(int64(limit)))
	if err != nil {
		return nil, err
	}
	err = cursor.All(ctx, &artworks)
	if err != nil {
		return nil, err
	}
	return artworks, nil
}

func GetArtworksByArtistUsername(ctx context.Context, artistUsername string, limit int) ([]*model.ArtworkModel, error) {
	var artworks []*model.ArtworkModel
	artists, err := GetArtistsByUserNameLike(ctx, artistUsername)
	if err != nil {
		return nil, err
	}
	artistIDs := make([]primitive.ObjectID, len(artists))
	for i, artist := range artists {
		artistIDs[i] = artist.ID
	}
	cursor, err := artworkCollection.Find(ctx, bson.M{"artist_id": bson.M{"$in": artistIDs}}, options.Find().SetLimit(int64(limit)))
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

func UpdateArtworkR18ByID(ctx context.Context, id primitive.ObjectID, r18 bool) (*mongo.UpdateResult, error) {
	return artworkCollection.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": bson.M{"r18": r18}})
}

func DeleteArtworkByID(ctx context.Context, id primitive.ObjectID) (*mongo.DeleteResult, error) {
	return artworkCollection.DeleteOne(ctx, bson.M{"_id": id})
}

func DeleteArtworkPicturesByID(ctx context.Context, artworkID primitive.ObjectID, pictureIDs []primitive.ObjectID) (*mongo.UpdateResult, error) {
	return artworkCollection.UpdateOne(ctx, bson.M{"_id": artworkID}, bson.M{"$pull": bson.M{"pictures": bson.M{"$in": pictureIDs}}})
}
