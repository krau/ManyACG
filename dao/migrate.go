package dao

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
)

func AddLikeCountToArtwork(ctx context.Context) error {
	_, err := artworkCollection.UpdateMany(ctx, bson.M{"like_count": bson.M{"$exists": false}}, bson.M{"$set": bson.M{"like_count": int64(0)}})
	return err
}
