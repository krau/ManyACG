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
	"encoding/json"
	"fmt"
	"path/filepath"
	"time"

	"github.com/imroc/req/v3"
	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegoutil"
	"go.mongodb.org/mongo-driver/bson"
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

func FixTwitterArtists(ctx context.Context, bot *telego.Bot, message *telego.Message) {
	client := req.C().ImpersonateChrome().SetCommonRetryCount(3).
		SetCommonRetryBackoffInterval(1*time.Second, 5*time.Second).
		SetCommonRetryFixedInterval(2 * time.Second).
		EnableDebugLog()
	if config.Cfg.Source.Proxy != "" {
		client.SetProxyURL(config.Cfg.Source.Proxy)
	}
	sendMessage := bot != nil && message != nil

	collection := dao.DB.Collection("Artists")
	if collection == nil {
		Logger.Errorf("Failed to get collection")
		if sendMessage {
			bot.SendMessage(telegoutil.Messagef(
				message.Chat.ChatID(),
				"Failed to get collection",
			))
		}
		return
	}
	total, err := collection.CountDocuments(ctx, bson.M{"type": "twitter"})
	if err != nil {
		Logger.Errorf("Failed to count artists: %v", err)
		if sendMessage {
			bot.SendMessage(telegoutil.Messagef(
				message.Chat.ChatID(),
				"Failed to count artists: %s",
				err.Error(),
			))
		}
		return
	}
	Logger.Infof("Found %d artists", total)
	if sendMessage {
		bot.SendMessage(telegoutil.Messagef(
			message.Chat.ChatID(),
			"Found %d artists",
			total,
		))
	}
	cursor, err := collection.Find(ctx, bson.M{"type": "twitter"})
	if err != nil {
		Logger.Errorf("Failed to find artists: %v", err)
		if sendMessage {
			bot.SendMessage(telegoutil.Messagef(
				message.Chat.ChatID(),
				"Failed to find artists: %s",
				err.Error(),
			))
		}
		return
	}
	defer cursor.Close(ctx)
	apiBase := fmt.Sprintf("https://api.%s/", config.Cfg.Source.Twitter.FxTwitterDomain)
	type ArtistResp struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		User    struct {
			ID         string `json:"id"`
			Name       string `json:"name"`
			ScreenName string `json:"screen_name"`
		} `json:"user"`
	}

	// 创建一个集合，用于存储失败文档id
	if err := dao.DB.CreateCollection(ctx, "failedArtists"); err != nil {
		Logger.Errorf("Failed to create collection: %v", err)
	}
	failedCollection := dao.DB.Collection("failedArtists")
	failed, count := 0, 0
	for cursor.Next(ctx) {
		count++
		var artist model.ArtistModel
		if err := cursor.Decode(&artist); err != nil {
			Logger.Errorf("Failed to decode artist: %v", err)
			failed++
			continue
		}
		updateArtist := func() error {
			resp, err := client.R().Get(apiBase + artist.Username)
			if err != nil {
				Logger.Errorf("Failed to get artist: %v", err)
				return err
			}
			var artistResp ArtistResp
			if err := json.Unmarshal(resp.Bytes(), &artistResp); err != nil {
				Logger.Errorf("Failed to unmarshal artist: %v", err)
				return err
			}
			if artistResp.Code != 200 {
				Logger.Errorf("Failed to get artist: %v", artistResp.Message)
				return fmt.Errorf("failed to get artist: %s", artistResp.Message)
			}
			artist.UID = artistResp.User.ID
			artist.Name = artistResp.User.Name
			if _, err := dao.UpdateArtist(ctx, &artist); err != nil {
				Logger.Errorf("Failed to update artist: %v", err)
				return err
			}
			return nil
		}
		if err := updateArtist(); err != nil {
			if _, err := failedCollection.InsertOne(ctx, artist); err != nil {
				Logger.Errorf("Failed to insert failed artist: %v", err)
			}
			failed++
		}
		time.Sleep(1 * time.Second)
	}
	Logger.Infof("Processed %d artists, %d failed", count, failed)
	if sendMessage {
		bot.SendMessage(telegoutil.Messagef(
			message.Chat.ChatID(),
			"Processed %d artists, %d failed",
			count,
			failed,
		))
	}
}

func Migrate(ctx context.Context) {
	Logger.Noticef("Starting migration")

	// Logger.Infof("Adding likes field")
	// if err := dao.AddLikeCountToArtwork(ctx); err != nil {
	// 	Logger.Errorf("Failed to add likes field: %v", err)
	// }

	// Logger.Infof("Migrating storage info")
	// if err := dao.MigrateStorageInfo(ctx); err != nil {
	// 	Logger.Errorf("Failed to migrate storage info: %v", err)
	// }

	// Logger.Infof("Tidying artist")
	// if err := dao.TidyArtist(ctx); err != nil {
	// 	Logger.Errorf("Failed to tidy artist: %v", err)
	// }

	Logger.Infof("Convert artist uid to string")
	if err := dao.ConvertArtistUIDToString(ctx); err != nil {
		Logger.Errorf("Failed to convert artist uid to string: %v", err)
	}

	Logger.Noticef("Migration completed")
}
