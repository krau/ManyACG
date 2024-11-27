package service

import (
	"context"
	"fmt"
	"image"
	"io"

	"github.com/krau/ManyACG/common"
	"github.com/krau/ManyACG/dao"
	"github.com/krau/ManyACG/storage"
	"github.com/krau/ManyACG/types"

	"github.com/duke-git/lancet/v2/slice"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// Deprecated: MessageID 现在可能为 0
func GetPictureByMessageID(ctx context.Context, messageID int) (*types.Picture, error) {
	pictureModel, err := dao.GetPictureByMessageID(ctx, messageID)
	if err != nil {
		return nil, err
	}
	return pictureModel.ToPicture(), nil
}

func GetPictureByID(ctx context.Context, id primitive.ObjectID) (*types.Picture, error) {
	pictureModel, err := dao.GetPictureByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return pictureModel.ToPicture(), nil
}

func GetRandomPictures(ctx context.Context, limit int) ([]*types.Picture, error) {
	pictures, err := dao.GetRandomPictures(ctx, limit)
	if err != nil {
		return nil, err
	}
	var result []*types.Picture
	for _, picture := range pictures {
		result = append(result, picture.ToPicture())
	}
	return result, nil
}

func UpdatePictureTelegramInfo(ctx context.Context, picture *types.Picture, telegramInfo *types.TelegramInfo) error {
	pictureModel, err := dao.GetPictureByOriginal(ctx, picture.Original)
	if err != nil {
		return err
	}
	_, err = dao.UpdatePictureTelegramInfoByID(ctx, pictureModel.ID, telegramInfo)
	if err != nil {
		return err
	}
	return nil
}

/*
通过消息删除 Picture

如果删除后 Artwork 中没有 Picture , 则也删除 Artwork

不会对存储进行操作
*/
// Deprecated: MessageID 现在不唯一且可能为 0
func DeletePictureByMessageID(ctx context.Context, messageID int) error {
	pictureModel, err := dao.GetPictureByMessageID(ctx, messageID)
	if err != nil {
		return err
	}
	session, err := dao.Client.StartSession()
	if err != nil {
		return err
	}
	defer session.EndSession(ctx)
	_, err = session.WithTransaction(ctx, func(ctx mongo.SessionContext) (interface{}, error) {
		_, err = dao.DeletePicturesByArtworkID(ctx, pictureModel.ArtworkID)
		if err != nil {
			return nil, err
		}
		artworkModel, err := dao.GetArtworkByID(ctx, pictureModel.ArtworkID)
		if err != nil {
			return nil, err
		}
		if len(artworkModel.Pictures) == 0 {
			_, err := dao.DeleteArtworkByID(ctx, pictureModel.ArtworkID)
			if err != nil {
				return nil, err
			}
			_, err = dao.CreateDeleted(ctx, &types.DeletedModel{
				SourceURL: artworkModel.SourceURL,
				ArtworkID: artworkModel.ID,
			})
			if err != nil {
				return nil, err
			}
		}
		return nil, nil
	}, options.Transaction().SetReadPreference(readpref.Primary()))
	if err != nil {
		return err
	}
	return nil
}

// 删除单张图片, 如果删除后对应的 artwork 中没有图片, 则也删除 artwork
//
// 删除后对 artwork 的 pictures 的 index 进行重整
func DeletePictureByID(ctx context.Context, id primitive.ObjectID) error {
	toDeletePictureModel, err := dao.GetPictureByID(ctx, id)
	if err != nil {
		return err
	}
	artworkModel, err := dao.GetArtworkByID(ctx, toDeletePictureModel.ArtworkID)
	if err != nil {
		return err
	}
	session, err := dao.Client.StartSession()
	if err != nil {
		return err
	}
	defer session.EndSession(ctx)
	_, err = session.WithTransaction(ctx, func(ctx mongo.SessionContext) (any, error) {
		if len(artworkModel.Pictures) == 1 {
			err := deleteArtwork(ctx, artworkModel.ID, artworkModel.SourceURL)
			return nil, err
		}

		_, err := dao.DeletePictureByID(ctx, id)
		if err != nil {
			return nil, err
		}

		newPictureIDs := slice.Filter(artworkModel.Pictures, func(index int, item primitive.ObjectID) bool {
			return item.Hex() != toDeletePictureModel.ID.Hex()
		})

		if _, err := dao.UpdateArtworkPicturesByID(ctx, artworkModel.ID, newPictureIDs); err != nil {
			return nil, err
		}
		return nil, TidyArtworkPictureIndexByID(ctx, artworkModel.ID)
	}, options.Transaction().SetReadPreference(readpref.Primary()))
	return err
}

func GetPicturesByHashHammingDistance(ctx context.Context, hash string, distance int) ([]*types.Picture, error) {
	if hash == "" {
		return nil, nil
	}
	pictures, err := dao.GetPicturesByHashHammingDistance(ctx, hash, distance)
	if err != nil {
		return nil, err
	}
	var result []*types.Picture
	for _, picture := range pictures {
		result = append(result, picture.ToPicture())
	}
	return result, nil
}

func ProcessPictureHashAndUpdate(ctx context.Context, picture *types.Picture) error {
	pictureModel, err := dao.GetPictureByOriginal(ctx, picture.Original)
	if err != nil {
		return err
	}
	var file io.ReadCloser
	if picture.StorageInfo.Original != nil {
		file, err = storage.GetFileStream(ctx, picture.StorageInfo.Original)
	} else {
		file, err = common.GetBodyReader(ctx, picture.Original, nil)
	}
	if err != nil {
		return err
	}
	defer file.Close()
	img, _, err := image.Decode(file)
	if err != nil {
		return fmt.Errorf("failed to decode image: %w", err)
	}

	hash, err := common.GetImagePhash(img)
	if err != nil {
		return err
	}
	_, err = dao.UpdatePictureHashByID(ctx, pictureModel.ID, hash)
	if err != nil {
		return err
	}
	if picture.Width == 0 || picture.Height == 0 {
		width, height, err := common.GetImageSize(img)
		if err != nil {
			return err
		}
		_, err = dao.UpdatePictureSizeByID(ctx, pictureModel.ID, width, height)
		if err != nil {
			return err
		}
	}
	return nil
}

type processPictureTask struct {
	Picture *types.Picture
	Ctx     context.Context
}

var processPictureTaskChan = make(chan *processPictureTask)

func AddProcessPictureTask(ctx context.Context, picture *types.Picture) {
	processPictureTaskChan <- &processPictureTask{
		Picture: picture,
		Ctx:     ctx,
	}
}

func listenProcessPictureTask() {
	for task := range processPictureTaskChan {
		err := ProcessPictureHashAndUpdate(task.Ctx, task.Picture)
		if err != nil {
			common.Logger.Errorf("error when processing picture %s: %s", task.Picture.Original, err)
		}
	}
}
