package service

import (
	"ManyACG/common"
	"ManyACG/dao"
	"ManyACG/dao/model"
	"ManyACG/storage"
	"ManyACG/types"
	"context"

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
		_, err = dao.DeleteArtworkPicturesByID(ctx, pictureModel.ArtworkID, []primitive.ObjectID{pictureModel.ID})
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

func ProcessPictureAndUpdate(ctx context.Context, picture *types.Picture) error {
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
		fileBytes, err := storage.GetStorage().GetFile(picture.StorageInfo)
		if err != nil {
			return nil, err
		}
		hash, err := common.GetPhash(fileBytes)
		if err != nil {
			return nil, err
		}
		blurscore, err := common.GetBlurScore(fileBytes)
		if err != nil {
			return nil, err
		}
		_, err = dao.UpdatePictureHashAndBlurScoreByID(ctx, pictureModel.ID, hash, blurscore)
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
