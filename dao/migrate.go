// 此文件内所有操作应停机执行

package dao

import (
	"context"
	"fmt"

	"github.com/krau/ManyACG/types"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
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

func TidyArtist(ctx context.Context) error {
	// 清理没有任何 artwork 的 artist (遍历 artists )
	if err := cleanNoArtworkArtists(ctx); err != nil {
		return fmt.Errorf("failed to clean no artwork artists: %w", err)
	}
	// 通过 source type 和 username 合并相同的 artist, 同时更改对应的 artwork 为合并后的同一个
	if err := mergeDupArtist(ctx); err != nil {
		return fmt.Errorf("failed to merge duplicate artists: %w", err)
	}

	return nil
}

func cleanNoArtworkArtists(ctx context.Context) error {
	fmt.Println("Cleaning artists without any artwork")
	cursor, err := artistCollection.Find(ctx, bson.M{})
	if err != nil {
		return fmt.Errorf("failed to find artists: %w", err)
	}
	defer cursor.Close(ctx)
	var artistsToDelete []primitive.ObjectID
	for cursor.Next(ctx) {
		var artist types.ArtistModel
		if err := cursor.Decode(&artist); err != nil {
			return fmt.Errorf("failed to decode artist: %w", err)
		}
		count, err := artworkCollection.CountDocuments(ctx, bson.M{"artist_id": artist.ID})
		if err != nil {
			return fmt.Errorf("failed to count artwork: %w", err)
		}

		if count == 0 {
			artistsToDelete = append(artistsToDelete, artist.ID)
		}
	}

	if len(artistsToDelete) == 0 {
		fmt.Println("No artist to delete")
		return nil
	}

	res, err := artistCollection.DeleteMany(ctx, bson.M{"_id": bson.M{"$in": artistsToDelete}})
	if err != nil {
		return fmt.Errorf("failed to delete artists: %w", err)
	}
	fmt.Printf("Deleted %d artists\n", res.DeletedCount)
	return nil
}

func mergeDupArtist(ctx context.Context) error {
	fmt.Println("Merging duplicate artists")
	pipeline := mongo.Pipeline{
		{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: bson.D{{Key: "type", Value: "$type"}, {Key: "username", Value: "$username"}}},
			{Key: "ids", Value: bson.D{{Key: "$push", Value: "$_id"}}},
			{Key: "first", Value: bson.D{{Key: "$first", Value: "$$ROOT"}}},
			{Key: "count", Value: bson.D{{Key: "$sum", Value: 1}}},
		}}},
		{{Key: "$match", Value: bson.D{
			{Key: "count", Value: bson.D{{Key: "$gt", Value: 1}}},
		}}},
	}
	cursor, err := artistCollection.Aggregate(ctx, pipeline)
	if err != nil {
		return fmt.Errorf("failed to aggregate artists: %w", err)
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var result struct {
			ID    bson.M               `bson:"_id"`
			IDs   []primitive.ObjectID `bson:"ids"`
			First types.ArtistModel    `bson:"first"`
			Count int                  `bson:"count"`
		}
		if err := cursor.Decode(&result); err != nil {
			return fmt.Errorf("failed to decode result: %w", err)
		}

		mainArtistID := result.First.ID

		artistsToDelete := result.IDs[1:]
		res, err := artworkCollection.UpdateMany(ctx, bson.M{"artist_id": bson.M{"$in": artistsToDelete}}, bson.M{"$set": bson.M{"artist_id": mainArtistID}})
		if err != nil {
			return fmt.Errorf("failed to update artworks: %w", err)
		}
		fmt.Printf("Updated %d artworks\n", res.ModifiedCount)

		delRes, err := artistCollection.DeleteMany(ctx, bson.M{"_id": bson.M{"$in": artistsToDelete}})
		if err != nil {
			return fmt.Errorf("failed to delete artists: %w", err)
		}
		fmt.Printf("Deleted %d artists\n", delRes.DeletedCount)
		fmt.Printf("Merged %d artists into %s\n", result.Count, result.First.Username)
	}

	return nil
}

func ConvertArtistUIDToString(ctx context.Context) error {
	pipeline := []bson.M{
		{
			"$set": bson.M{
				"uid": bson.M{
					"$toString": "$uid",
				},
			},
		},
	}
	_, err := artistCollection.UpdateMany(ctx, bson.M{}, pipeline)
	return err
}

func AddAliasToAllTags(ctx context.Context) error {
	// 为所有 tag 添加 alias 属性, 默认值为包含其 name 的数组
	// 用于从 v0.64.1 及之前的版本迁移
	_, err := tagCollection.UpdateMany(
		ctx,
		bson.M{"alias": bson.M{"$exists": false}},
		bson.A{
			bson.M{
				"$set": bson.M{
					"alias": bson.M{
						"$ifNull": bson.A{
							"$alias",
							bson.A{"$name"},
						},
					},
				},
			},
		},
	)
	return err
}

func TidyTag(ctx context.Context) error {
	fmt.Println("正在清理未使用的标签")

	pipeline := mongo.Pipeline{
		{{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "Artworks"},
			{Key: "localField", Value: "_id"},
			{Key: "foreignField", Value: "tags"},
			{Key: "as", Value: "artwork_tags"},
		}}},
		{{Key: "$match", Value: bson.D{
			{Key: "artwork_tags", Value: bson.D{{Key: "$eq", Value: bson.A{}}}},
		}}},
		{{Key: "$project", Value: bson.D{
			{Key: "_id", Value: 1},
			{Key: "name", Value: 1},
			{Key: "alias", Value: 1},
		}}},
	}

	cursor, err := tagCollection.Aggregate(ctx, pipeline)
	if err != nil {
		return fmt.Errorf("查找未使用的标签失败: %w", err)
	}
	defer cursor.Close(ctx)

	var tagsToDelete []primitive.ObjectID
	for cursor.Next(ctx) {
		var result struct {
			ID    primitive.ObjectID `bson:"_id"`
			Name  string             `bson:"name"`
			Alias []string           `bson:"alias"`
		}
		if err := cursor.Decode(&result); err != nil {
			return fmt.Errorf("解析标签数据失败: %w", err)
		}
		tagsToDelete = append(tagsToDelete, result.ID)
		fmt.Printf("将删除未使用的标签: %s (别名: %v)\n", result.Name, result.Alias)
	}

	if len(tagsToDelete) == 0 {
		fmt.Println("没有需要删除的标签")
		return nil
	}

	res, err := tagCollection.DeleteMany(ctx, bson.M{"_id": bson.M{"$in": tagsToDelete}})
	if err != nil {
		return fmt.Errorf("删除未使用的标签失败: %w", err)
	}

	fmt.Printf("总共删除了 %d 个未使用的标签\n", res.DeletedCount)
	return nil
}
