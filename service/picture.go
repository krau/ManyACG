package service

import (
	"ManyACG-Bot/dao"
	"ManyACG-Bot/dao/model"
	"ManyACG-Bot/types"
	"context"

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
