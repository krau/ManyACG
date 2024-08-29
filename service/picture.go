package service

import (
	"ManyACG/common"
	"ManyACG/dao"
	"ManyACG/model"
	"ManyACG/storage"
	"ManyACG/types"
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
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

func UpdatePictureTelegramInfo(ctx context.Context, picture *types.Picture, telegramInfo *types.TelegramInfo) error {
	pictureModel, err := dao.GetPictureByOriginal(ctx, picture.Original)
	if err != nil {
		return err
	}
	session, err := dao.Client.StartSession()
	if err != nil {
		return err
	}
	defer session.EndSession(ctx)
	_, err = session.WithTransaction(ctx, func(ctx mongo.SessionContext) (interface{}, error) {
		_, err := dao.UpdatePictureTelegramInfoByID(ctx, pictureModel.ID, telegramInfo)
		if err != nil {
			return nil, err
		}
		return nil, nil
	})
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
		_, err = dao.DeletePicturesByIDs(ctx, []primitive.ObjectID{pictureModel.ID})
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
			_, err = dao.CreateDeleted(ctx, &model.DeletedModel{
				SourceURL: artworkModel.SourceURL,
				ArtworkID: artworkModel.ID,
			})
			if err != nil {
				return nil, err
			}
		}
		return nil, nil
	})
	if err != nil {
		return err
	}
	return nil
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

func ProcessPictureHashAndSizeAndUpdate(ctx context.Context, picture *types.Picture) error {
	pictureModel, err := dao.GetPictureByOriginal(ctx, picture.Original)
	if err != nil {
		return err
	}
	session, err := dao.Client.StartSession()
	if err != nil {
		return err
	}
	defer session.EndSession(ctx)
	_, err = session.WithTransaction(ctx, func(ctx mongo.SessionContext) (interface{}, error) {
		fileBytes, err := storage.GetFile(ctx, picture.StorageInfo.Original)
		if err != nil {
			return nil, err
		}
		defer func() {
			fileBytes = nil
		}()
		hash, err := common.GetImagePhash(fileBytes)
		if err != nil {
			return nil, err
		}
		blurscore, err := common.GetImageBlurScore(fileBytes)
		if err != nil {
			return nil, err
		}
		_, err = dao.UpdatePictureHashAndBlurScoreByID(ctx, pictureModel.ID, hash, blurscore)
		if err != nil {
			return nil, err
		}
		if picture.Width == 0 || picture.Height == 0 {
			width, height, err := common.GetImageSize(fileBytes)
			if err != nil {
				return nil, err
			}
			_, err = dao.UpdatePictureSizeByID(ctx, pictureModel.ID, width, height)
			if err != nil {
				return nil, err
			}
		}
		return nil, nil
	})
	if err != nil {
		return err
	}
	return nil
}
