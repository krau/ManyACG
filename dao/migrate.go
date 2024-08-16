package dao

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
)

func AddLikesField(ctx context.Context) error {
	_, err := artworkCollection.UpdateMany(ctx, bson.M{}, bson.M{"$set": bson.M{"likes": 0}})
	return err
}
