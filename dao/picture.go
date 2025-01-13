package dao

import (
	"context"

	"github.com/krau/ManyACG/types"

	"github.com/corona10/goimagehash"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var pictureCollection *mongo.Collection

func CreatePicture(ctx context.Context, picture *types.PictureModel) (*mongo.InsertOneResult, error) {
	return pictureCollection.InsertOne(ctx, picture)
}

func CreatePictures(ctx context.Context, pictures []*types.PictureModel) (*mongo.InsertManyResult, error) {
	var docs []interface{}
	for _, picture := range pictures {
		if picture.TelegramInfo == nil {
			picture.TelegramInfo = &types.TelegramInfo{}
		}
		docs = append(docs, picture)
	}
	return pictureCollection.InsertMany(ctx, docs)
}

// Deprecated: MessageID 现在可能为 0
func GetPictureByMessageID(ctx context.Context, messageID int) (*types.PictureModel, error) {
	var picture types.PictureModel
	err := pictureCollection.FindOne(ctx, bson.M{"telegram_info.message_id": messageID}).Decode(&picture)
	if err != nil {
		return nil, err
	}
	return &picture, nil
}

func GetPictureByOriginal(ctx context.Context, original string) (*types.PictureModel, error) {
	var picture types.PictureModel
	err := pictureCollection.FindOne(ctx, bson.M{"original": original}).Decode(&picture)
	if err != nil {
		return nil, err
	}
	return &picture, nil
}

func GetPictureByID(ctx context.Context, id primitive.ObjectID) (*types.PictureModel, error) {
	var picture types.PictureModel
	err := pictureCollection.FindOne(ctx, bson.M{"_id": id}).Decode(&picture)
	if err != nil {
		return nil, err
	}
	return &picture, nil
}

func GetPicturesByHash(ctx context.Context, hash string) ([]*types.PictureModel, error) {
	cursor, err := pictureCollection.Find(ctx, bson.M{"hash": hash})
	if err != nil {
		return nil, err
	}
	var pictures []*types.PictureModel
	err = cursor.All(ctx, &pictures)
	if err != nil {
		return nil, err
	}
	return pictures, nil
}

func GetRandomPictures(ctx context.Context, limit int) ([]*types.PictureModel, error) {
	cursor, err := pictureCollection.Aggregate(ctx, mongo.Pipeline{
		bson.D{{Key: "$sample", Value: bson.M{"size": limit}}},
	})
	if err != nil {
		return nil, err
	}
	var pictures []*types.PictureModel
	err = cursor.All(ctx, &pictures)
	if err != nil {
		return nil, err
	}
	return pictures, nil
}

/*
全库遍历搜索
*/
func GetPicturesByHashHammingDistance(ctx context.Context, hashStr string, distance int) ([]*types.PictureModel, error) {
	hashToCompare, err := goimagehash.ImageHashFromString(hashStr)
	if err != nil {
		return nil, err
	}
	filter := bson.M{
		"hash": bson.M{"$ne": ""},
	}
	cursor, err := pictureCollection.Find(ctx, filter, options.Find().SetProjection(bson.M{"hash": 1, "artwork_id": 1, "index": 1, "telegram_info": 1}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var pictures []*types.PictureModel
	for cursor.Next(ctx) {
		var picture types.PictureModel
		err = cursor.Decode(&picture)
		if err != nil {
			return nil, err
		}

		hash, err := goimagehash.ImageHashFromString(picture.Hash)
		if err != nil {
			return nil, err
		}

		dist, err := hash.Distance(hashToCompare)
		if err != nil {
			return nil, err
		}

		if dist <= distance {
			pictures = append(pictures, &picture)
		}
	}

	return pictures, nil
}

func GetNoHashPictures(ctx context.Context) ([]*types.PictureModel, error) {
	cursor, err := pictureCollection.Find(ctx, bson.M{"hash": ""})
	if err != nil {
		return nil, err
	}
	var pictures []*types.PictureModel
	err = cursor.All(ctx, &pictures)
	if err != nil {
		return nil, err
	}
	return pictures, nil
}

func GetNoRegularAndThumbPictures(ctx context.Context) ([]*types.PictureModel, error) {
	cursor, err := pictureCollection.Find(ctx, bson.M{"storage_info.regular": nil, "storage_info.thumb": nil})
	if err != nil {
		return nil, err
	}
	var pictures []*types.PictureModel
	err = cursor.All(ctx, &pictures)
	if err != nil {
		return nil, err
	}
	return pictures, nil
}

func GetPictureCount(ctx context.Context) (int64, error) {
	return pictureCollection.CountDocuments(ctx, bson.M{})
}

func UpdatePictureIndexByID(ctx context.Context, id primitive.ObjectID, index uint) (*mongo.UpdateResult, error) {
	return pictureCollection.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": bson.M{"index": index}})
}

func UpdatePictureTelegramInfoByID(ctx context.Context, id primitive.ObjectID, telegramInfo *types.TelegramInfo) (*mongo.UpdateResult, error) {
	return pictureCollection.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": bson.M{"telegram_info": telegramInfo}})
}

func UpdatePictureHashByID(ctx context.Context, id primitive.ObjectID, hash string) (*mongo.UpdateResult, error) {
	return pictureCollection.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": bson.M{"hash": hash}})
}

func UpdatePictureSizeByID(ctx context.Context, id primitive.ObjectID, width, height int) (*mongo.UpdateResult, error) {
	return pictureCollection.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": bson.M{"width": width, "height": height}})
}

func UpdatePictureByID(ctx context.Context, id primitive.ObjectID, picture *types.PictureModel) (*mongo.UpdateResult, error) {
	return pictureCollection.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": picture})
}

func UpdatePictureStorageInfoByID(ctx context.Context, id primitive.ObjectID, storageInfo *types.StorageInfo) (*mongo.UpdateResult, error) {
	return pictureCollection.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": bson.M{"storage_info": storageInfo}})
}

func DeletePicturesByArtworkID(ctx context.Context, artworkID primitive.ObjectID) (*mongo.DeleteResult, error) {
	return pictureCollection.DeleteMany(ctx, bson.M{"artwork_id": artworkID})
}

func DeletePictureByID(ctx context.Context, id primitive.ObjectID) (*mongo.DeleteResult, error) {
	return pictureCollection.DeleteOne(ctx, bson.M{"_id": id})
}
