package dao

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
)

func AddLikeCountToArtwork(ctx context.Context) error {
	_, err := artworkCollection.UpdateMany(ctx, bson.M{"like_count": bson.M{"$exists": false}}, bson.M{"$set": bson.M{"like_count": int64(0)}})
	return err
}

func MigrateStorageInfo(ctx context.Context) error {
	/*
		原先:
			Picture: {
				storage_info: {
					type: string,
					path: string,
				}
			}

		现在:
			Picture: {
				storage_info: {
					original (原先的 storage_info): {
						type: string,
						path: string,
					},
					regular: {
						type: string,
						path: string,
					},
					thumb: {
						type: string,
						path: string,
					},
				}
			}
	*/

	// Step 1: Restructure existing storage_info
	_, err := pictureCollection.UpdateMany(
		ctx,
		bson.M{"storage_info": bson.M{"$exists": true, "$type": "object"}},
		bson.A{
			bson.M{"$set": bson.M{
				"storage_info": bson.M{
					"original": "$storage_info",
					"regular":  nil,
					"thumb":    nil,
				},
			}},
		},
	)
	if err != nil {
		return fmt.Errorf("failed to restructure existing storage_info: %w", err)
	}

	// Step 2: Remove storage_info.type and storage_info.path
	_, err = pictureCollection.UpdateMany(
		ctx,
		bson.M{"storage_info.original.type": bson.M{"$exists": true}},
		bson.M{"$unset": bson.M{
			"storage_info.type": "",
			"storage_info.path": "",
		}},
	)
	if err != nil {
		return fmt.Errorf("failed to remove storage_info.type and storage_info.path: %w", err)
	}

	// Step 3: Handle documents without storage_info or with incorrect structure
	_, err = pictureCollection.UpdateMany(
		ctx,
		bson.M{"$or": []bson.M{
			{"storage_info": bson.M{"$exists": false}},
			{"storage_info.original": bson.M{"$exists": false}},
		}},
		bson.M{"$set": bson.M{"storage_info": bson.M{
			"original": bson.M{
				"type": nil,
				"path": nil,
			},
			"regular": nil,
			"thumb":   nil,
		}}},
	)
	if err != nil {
		return fmt.Errorf("failed to handle documents without proper storage_info: %w", err)
	}

	return nil

}
