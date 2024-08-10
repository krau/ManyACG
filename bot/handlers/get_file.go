package handlers

import (
	"ManyACG/common"
	"ManyACG/config"
	"ManyACG/service"
	"ManyACG/storage"
	"ManyACG/telegram"
	"ManyACG/types"
	"bytes"
	"context"
	"errors"
	"path/filepath"
	"strconv"
	"time"

	. "ManyACG/logger"

	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegoutil"
	"go.mongodb.org/mongo-driver/mongo"
)

func GetPictureFile(ctx context.Context, bot *telego.Bot, message telego.Message) {
	messageOrigin, ok := telegram.GetMessageOriginChannelArtworkPost(ctx, bot, message)
	if !ok {
		if message.ReplyToMessage == nil {
			telegram.ReplyMessage(bot, message, "请回复一条频道的图片消息")
			return
		}
		fileBytes, err := telegram.GetMessagePhotoFileBytes(bot, message.ReplyToMessage)
		if err != nil {
			telegram.ReplyMessage(bot, message, "请回复一条频道的图片消息")
			return
		}
		hash, err := common.GetImagePhash(fileBytes)
		if err != nil {
			telegram.ReplyMessage(bot, message, "请回复一条频道的图片消息")
			return
		}
		pictures, err := service.GetPicturesByHashHammingDistance(ctx, hash, 10)
		if err != nil || len(pictures) == 0 {
			telegram.ReplyMessage(bot, message, "请回复一条频道的图片消息")
			return
		}
		picture := pictures[0]
		_, err = telegram.SendPictureFileByMessageID(ctx, bot, message, picture.TelegramInfo.MessageID)
		if err != nil {
			telegram.ReplyMessage(bot, message, "文件发送失败: "+err.Error())
			return
		}
		return
	}
	pictureMessageID := messageOrigin.MessageID
	cmd, _, args := telegoutil.ParseCommand(message.Text)
	multiple := cmd == "files"
	artwork, err := service.GetArtworkByMessageID(ctx, pictureMessageID)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			telegram.ReplyMessage(bot, message, "这张图片未在数据库中呢")
			return
		}
		Logger.Errorf("获取作品失败: %s", err)
		telegram.ReplyMessage(bot, message, "获取失败, 去找管理员反馈吧~")
		return
	}

	if multiple {
		getArtworkFiles(ctx, bot, message, artwork)
		return
	}

	if len(args) > 0 {
		index, err := strconv.Atoi(args[0])
		if err == nil && index > 0 {
			if index > len(artwork.Pictures) {
				telegram.ReplyMessage(bot, message, "这个作品没有这么多图片")
				return
			}
			picture := artwork.Pictures[index-1]
			pictureMessageID = picture.TelegramInfo.MessageID
		}
	}
	_, err = telegram.SendPictureFileByMessageID(ctx, bot, message, pictureMessageID)
	if err != nil {
		telegram.ReplyMessage(bot, message, "文件发送失败: "+err.Error())
		return
	}
}

func getArtworkFiles(ctx context.Context, bot *telego.Bot, message telego.Message, artwork *types.Artwork) {
	defer func() {
		if r := recover(); r != nil {
			Logger.Fatalf("获取文件失败: %s", r)
		}
	}()
	for i, picture := range artwork.Pictures {
		var file telego.InputFile
		alreadyCached := picture.TelegramInfo.DocumentFileID != ""
		if alreadyCached {
			file = telegoutil.FileFromID(picture.TelegramInfo.DocumentFileID)
		} else {
			data, err := storage.GetStorage().GetFile(picture.StorageInfo)
			if err != nil {
				telegram.ReplyMessage(bot, message, "获取文件失败: "+err.Error())
				return
			}
			file = telegoutil.File(telegoutil.NameReader(bytes.NewReader(data), filepath.Base(picture.StorageInfo.Path)))
		}
		documentMessage, err := bot.SendDocument(telegoutil.Document(message.Chat.ChatID(), file).
			WithReplyParameters(&telego.ReplyParameters{
				MessageID: message.MessageID,
			}).WithCaption(artwork.Title + "_" + strconv.Itoa(i+1)))
		if err != nil {
			Logger.Errorf("发送文件失败: %s", err)
			bot.SendMessage(telegoutil.Messagef(
				message.Chat.ChatID(),
				"发送第 %d 张图片时失败",
				i+1,
			).WithReplyParameters(&telego.ReplyParameters{
				MessageID: message.MessageID,
			}))
			continue
		}
		if documentMessage != nil {
			picture.TelegramInfo.DocumentFileID = documentMessage.Document.FileID
			if service.UpdatePictureTelegramInfo(ctx, picture, picture.TelegramInfo) != nil {
				Logger.Warnf("更新图片信息失败: %s", err)
			}
			if alreadyCached {
				time.Sleep(time.Duration(config.Cfg.Telegram.Sleep) * time.Second)
			}
		}
	}
}
