package dao

import (
	"context"

	"ManyACG/model"

	"github.com/corona10/goimagehash"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var pictureCollection *mongo.Collection

func CreatePicture(ctx context.Context, picture *model.PictureModel) (*mongo.InsertOneResult, error) {
	return pictureCollection.InsertOne(ctx, picture)
}

func CreatePictures(ctx context.Context, pictures []*model.PictureModel) (*mongo.InsertManyResult, error) {
	var docs []interface{}
	for _, picture := range pictures {
		docs = append(docs, picture)
	}
	return pictureCollection.InsertMany(ctx, docs)
}

func GetPictureByMessageID(ctx context.Context, messageID int) (*model.PictureModel, error) {
	var picture model.PictureModel
	err := pictureCollection.FindOne(ctx, bson.M{"telegram_info.message_id": messageID}).Decode(&picture)
	if err != nil {
		return nil, err
	}
	return &picture, nil
}

func GetPictureByOriginal(ctx context.Context, original string) (*model.PictureModel, error) {
	var picture model.PictureModel
	err := pictureCollection.FindOne(ctx, bson.M{"original": original}).Decode(&picture)
	if err != nil {
		return nil, err
	}
	return &picture, nil
}

func GetPictureByID(ctx context.Context, id primitive.ObjectID) (*model.PictureModel, error) {
	var picture model.PictureModel
	err := pictureCollection.FindOne(ctx, bson.M{"_id": id}).Decode(&picture)
	if err != nil {
		return nil, err
	}
	return &picture, nil
}

func GetPicturesByHash(ctx context.Context, hash string) ([]*model.PictureModel, error) {
	cursor, err := pictureCollection.Find(ctx, bson.M{"hash": hash})
	if err != nil {
		return nil, err
	}
	var pictures []*model.PictureModel
	err = cursor.All(ctx, &pictures)
	if err != nil {
		return nil, err
	}
	return pictures, nil
}

/*
	全库遍历搜索

1k 个图像每次搜索在 i7-12700h 上耗时 7ms 左右
TODO: 优化搜索速度和内存占用
*/
func GetPicturesByHashHammingDistance(ctx context.Context, hashStr string, distance int) ([]*model.PictureModel, error) {
	filter := bson.M{
		"hash": bson.M{"$ne": ""},
	}
	cursor, err := pictureCollection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var pictures []*model.PictureModel
	for cursor.Next(ctx) {
		var picture model.PictureModel
		err = cursor.Decode(&picture)
		if err != nil {
			return nil, err
		}

		hash, err := goimagehash.ImageHashFromString(picture.Hash)
		if err != nil {
			return nil, err
		}

		hashToCompare, err := goimagehash.ImageHashFromString(hashStr)
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

func GetNotProcessedPictures(ctx context.Context) ([]*model.PictureModel, error) {
	cursor, err := pictureCollection.Find(ctx, bson.M{"hash": ""})
	if err != nil {
		return nil, err
	}
	var pictures []*model.PictureModel
	err = cursor.All(ctx, &pictures)
	if err != nil {
		return nil, err
	}
	return pictures, nil
}

func UpdatePictureTelegramInfoByID(ctx context.Context, id primitive.ObjectID, telegramInfo *model.TelegramInfo) (*mongo.UpdateResult, error) {
	return pictureCollection.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": bson.M{"telegram_info": telegramInfo}})
}

func UpdatePictureHashAndBlurScoreByID(ctx context.Context, id primitive.ObjectID, hash string, blurScore float64) (*mongo.UpdateResult, error) {
	return pictureCollection.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": bson.M{"hash": hash, "blur_score": blurScore}})
}

func UpdatePictureSizeByID(ctx context.Context, id primitive.ObjectID, width, height int) (*mongo.UpdateResult, error) {
	return pictureCollection.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": bson.M{"width": width, "height": height}})
}

func DeletePicturesByIDs(ctx context.Context, ids []primitive.ObjectID) (*mongo.DeleteResult, error) {
	return pictureCollection.DeleteMany(ctx, bson.M{"_id": bson.M{"$in": ids}})
}

func DeletePicturesByArtworkID(ctx context.Context, artworkID primitive.ObjectID) (*mongo.DeleteResult, error) {
	return pictureCollection.DeleteMany(ctx, bson.M{"artwork_id": artworkID})
}
