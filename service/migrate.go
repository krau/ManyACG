package service

import (
	"ManyACG/common"
	"ManyACG/config"
	"ManyACG/dao"
	. "ManyACG/logger"
	"ManyACG/model"
	"ManyACG/storage"
	"ManyACG/types"
	"context"
	"fmt"
	"path/filepath"

	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegoutil"
	"go.mongodb.org/mongo-driver/mongo"
)

func ProcessPicturesHashAndSizeAndUpdate(ctx context.Context, bot *telego.Bot, message *telego.Message) {
	pictures, err := dao.GetNoHashPictures(ctx)
	sendMessage := bot != nil && message != nil
	if err != nil {
		Logger.Errorf("Failed to get not processed pictures: %v", err)
		if sendMessage {
			bot.SendMessage(telegoutil.Messagef(
				message.Chat.ChatID(),
				"Failed to get not processed pictures: %s",
				err.Error(),
			))
		}
		return
	}
	if sendMessage {
		bot.SendMessage(telegoutil.Messagef(
			message.Chat.ChatID(),
			"Found %d not processed pictures",
			len(pictures),
		))
	}

	failed := 0
	for _, picture := range pictures {
		if err := ProcessPictureHashAndSizeAndUpdate(ctx, picture.ToPicture()); err != nil {
			failed++
		}
	}
	Logger.Infof("Processed %d pictures, %d failed", len(pictures)-failed, failed)
	if sendMessage {
		bot.SendMessage(telegoutil.Messagef(
			message.Chat.ChatID(),
			"Processed %d pictures, %d failed",
			len(pictures)-failed,
			failed,
		))
	}
}

func StoragePictureRegularAndThumbAndUpdate(ctx context.Context, picture *model.PictureModel) error {
	pictureModel, err := dao.GetPictureByID(ctx, picture.ID)
	if err != nil {
		return err
	}
	artwork, err := GetArtworkByID(ctx, pictureModel.ArtworkID)
	if err != nil {
		return err
	}
	session, err := dao.Client.StartSession()
	if err != nil {
		return err
	}
	defer session.EndSession(ctx)
	migrateDir := config.Cfg.Storage.CacheDir + "/migrate/"
	_, err = session.WithTransaction(ctx, func(ctx mongo.SessionContext) (interface{}, error) {
		fileBytes, err := storage.GetFile(ctx, picture.StorageInfo.Original)
		if err != nil {
			return nil, err
		}
		defer func() {
			fileBytes = nil
		}()
		originalPath := migrateDir + filepath.Base(picture.StorageInfo.Original.Path)
		if err := common.MkFile(originalPath, fileBytes); err != nil {
			return nil, err
		}
		defer func() {
			common.PurgeFile(originalPath)
		}()

		regularPath := migrateDir + picture.ID.Hex() + "_regular.webp"
		if err := common.CompressImageByFFmpeg(originalPath, regularPath, 2560, 75); err != nil {
			return nil, err
		}
		defer func() {
			common.PurgeFile(regularPath)
		}()
		basePath := fmt.Sprintf("%s/%s", artwork.SourceType, common.ReplaceFileNameInvalidChar(artwork.Artist.Username))
		regularStoragePath := fmt.Sprintf("/regular/%s/%s", basePath, picture.ID.Hex()+"_regular.webp")
		regularDetail, err := storage.Save(ctx, regularPath, regularStoragePath, types.StorageType(config.Cfg.Storage.RegularType))
		if err != nil {
			return nil, err
		}
		pictureModel.StorageInfo.Regular = regularDetail
		if _, err := dao.UpdatePictureStorageInfoByID(ctx, pictureModel.ID, pictureModel.StorageInfo); err != nil {
			return nil, err
		}

		thumbPath := migrateDir + picture.ID.Hex() + "_thumb.webp"
		if err := common.CompressImageByFFmpeg(originalPath, thumbPath, 500, 75); err != nil {
			return nil, err
		}
		defer func() {
			common.PurgeFile(thumbPath)
		}()
		thumbStoragePath := fmt.Sprintf("/thumb/%s/%s", basePath, picture.ID.Hex()+"_thumb.webp")
		thumbDetail, err := storage.Save(ctx, thumbPath, thumbStoragePath, types.StorageType(config.Cfg.Storage.ThumbType))
		if err != nil {
			return nil, err
		}
		pictureModel.StorageInfo.Thumb = thumbDetail
		if _, err := dao.UpdatePictureStorageInfoByID(ctx, pictureModel.ID, pictureModel.StorageInfo); err != nil {
			return nil, err
		}
		return nil, nil
	})
	if err != nil {
		return err
	}
	return nil
}

func StoragePicturesRegularAndThumbAndUpdate(ctx context.Context, bot *telego.Bot, message *telego.Message) {
	pictures, err := dao.GetNoRegularAndThumbPictures(ctx)
	sendMessage := bot != nil && message != nil
	if err != nil {
		Logger.Errorf("Failed to get no regular and thumb pictures: %v", err)
		if sendMessage {
			bot.SendMessage(telegoutil.Messagef(
				message.Chat.ChatID(),
				"Failed to get no regular and thumb pictures: %s",
				err.Error(),
			))
		}
		return
	}
	if sendMessage {
		bot.SendMessage(telegoutil.Messagef(
			message.Chat.ChatID(),
			"Found %d no regular and thumb pictures",
			len(pictures),
		))
	}

	failed := 0
	for _, picture := range pictures {
		if err := StoragePictureRegularAndThumbAndUpdate(ctx, picture); err != nil {
			Logger.Errorf("Failed to storage regular and thumb picture: %v", err)
			failed++
		}
	}
	Logger.Infof("Processed %d pictures, %d failed", len(pictures)-failed, failed)
	if sendMessage {
		bot.SendMessage(telegoutil.Messagef(
			message.Chat.ChatID(),
			"Processed %d pictures, %d failed",
			len(pictures)-failed,
			failed,
		))
	}
}

func Migrate(ctx context.Context) {
	Logger.Noticef("Starting migration")
	Logger.Infof("Adding likes field")
	if err := dao.AddLikeCountToArtwork(ctx); err != nil {
		Logger.Errorf("Failed to add likes field: %v", err)
	}
	Logger.Infof("Migrating storage info")
	if err := dao.MigrateStorageInfo(ctx); err != nil {
		Logger.Errorf("Failed to migrate storage info: %v", err)
	}
	Logger.Noticef("Migration completed")
}
