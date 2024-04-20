package service

import (
	"ManyACG-Bot/dao"
	"ManyACG-Bot/dao/model"
	"ManyACG-Bot/types"
	"context"

	. "ManyACG-Bot/logger"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func GetPictureByMessageID(ctx context.Context, messageID int) (*types.Picture, error) {
	pictureModel, err := dao.GetPictureByMessageID(ctx, messageID)
	if err != nil {
		return nil, err
	}
	return &types.Picture{
		Index:     pictureModel.Index,
		Original:  pictureModel.Original,
		Thumbnail: pictureModel.Thumbnail,

		Width:        pictureModel.Width,
		Height:       pictureModel.Height,
		Hash:         pictureModel.Hash,
		BlurScore:    pictureModel.BlurScore,
		TelegramInfo: (*types.TelegramInfo)(pictureModel.TelegramInfo),
		StorageInfo:  (*types.StorageInfo)(pictureModel.StorageInfo),
	}, nil
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
		_, err := dao.UpdatePictureTelegramInfoByID(ctx, pictureModel.ID, (*model.TelegramInfo)(telegramInfo))
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
	删除 Picture

如果删除后 Artwork 中没有 Picture , 则也删除 Artwork
不会对存储进行操作
*/
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
		artworkResult, err := dao.DeleteArtworkPicturesByID(ctx, pictureModel.ArtworkID, []primitive.ObjectID{pictureModel.ID})
		if err != nil {
			return nil, err
		}
		if artworkResult.MatchedCount == 0 {
			Logger.Warnf("DeletePictureByMessageID: MatchedCount == 0")
		}

		pictureResult, err := dao.DeletePicturesByIDs(ctx, []primitive.ObjectID{pictureModel.ID})
		if err != nil {
			return nil, err
		}
		if pictureResult.DeletedCount == 0 {
			Logger.Warnf("DeletePictureByMessageID: DeletedCount == 0")
		}

		artworkModel, err := dao.GetArtworkByID(ctx, pictureModel.ArtworkID)
		if err != nil {
			return nil, err
		}

		if len(artworkModel.Pictures) == 0 {
			deleteResult, err := dao.DeleteArtworkByID(ctx, pictureModel.ArtworkID)
			if err != nil {
				return nil, err
			}
			if deleteResult.DeletedCount == 0 {
				Logger.Warnf("DeletePictureByMessageID: DeleteArtworkByID: DeletedCount == 0")
			}
		}
		return nil, nil
	})
	if err != nil {
		return err
	}
	return nil
}
