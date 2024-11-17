package handlers

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strconv"
	"time"

	"github.com/krau/ManyACG/common"
	"github.com/krau/ManyACG/config"
	"github.com/krau/ManyACG/service"
	"github.com/krau/ManyACG/sources"
	"github.com/krau/ManyACG/storage"
	"github.com/krau/ManyACG/telegram/utils"
	"github.com/krau/ManyACG/types"

	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegoutil"
	"go.mongodb.org/mongo-driver/mongo"
)

func GetPictureFile(ctx context.Context, bot *telego.Bot, message telego.Message) {
	var sourceURL string
	if message.ReplyToMessage != nil {
		sourceURL = utils.FindSourceURLForMessage(message.ReplyToMessage)
	} else {
		sourceURL = sources.FindSourceURL(message.Text)
	}
	if sourceURL == "" {
		getPictureByHash := func() *types.Picture {
			if message.ReplyToMessage == nil {
				return nil
			}
			fileBytes, err := utils.GetMessagePhotoFileBytes(bot, message.ReplyToMessage)
			if err != nil {
				return nil
			}
			hash, err := common.GetImagePhash(fileBytes)
			if err != nil {
				return nil
			}
			pictures, err := service.GetPicturesByHashHammingDistance(ctx, hash, 10)
			if err != nil || len(pictures) == 0 {
				return nil
			}
			return pictures[0]
		}
		picture := getPictureByHash()
		if picture == nil {
			utils.ReplyMessage(bot, message, "请回复一条频道的图片消息")
			return
		}
		_, err := utils.SendPictureFileByID(ctx, bot, message, ChannelChatID, picture.ID)
		if err != nil {
			utils.ReplyMessage(bot, message, "文件发送失败: "+err.Error())
			return
		}
		return
	}
	artwork, err := service.GetArtworkByURL(ctx, sourceURL)
	if err != nil {
		common.Logger.Errorf("获取作品信息失败: %s", err)
		if errors.Is(err, mongo.ErrNoDocuments) {
			utils.ReplyMessage(bot, message, "这张图片未在数据库中呢")
			return
		}
		utils.ReplyMessage(bot, message, "获取作品信息失败")
		return
	}
	cmd, _, args := telegoutil.ParseCommand(message.Text)
	multiple := cmd == "files"
	if multiple {
		getArtworkFiles(ctx, bot, message, artwork)
		return
	}
	var index int
	if len(args) != 0 {
		index, err = strconv.Atoi(args[0])
		if err != nil || index <= 0 {
			utils.ReplyMessage(bot, message, "请输入正确的作品序号, 从 1 开始")
			return
		}
		index--
	}
	if index > len(artwork.Pictures) {
		utils.ReplyMessage(bot, message, "这个作品没有这么多图片")
		return
	}
	picture := artwork.Pictures[index]
	_, err = utils.SendPictureFileByID(ctx, bot, message, ChannelChatID, picture.ID)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			utils.ReplyMessage(bot, message, "这张图片未在数据库中呢")
			return
		}
		common.Logger.Errorf("发送文件失败: %s", err)
		utils.ReplyMessage(bot, message, "发送文件失败, 去找管理员反馈吧~")
		return
	}
}

func getArtworkFiles(ctx context.Context, bot *telego.Bot, message telego.Message, artwork *types.Artwork) {
	defer func() {
		if r := recover(); r != nil {
			common.Logger.Fatalf("获取文件失败: %s", r)
		}
	}()
	for i, picture := range artwork.Pictures {
		var file telego.InputFile
		alreadyCached := picture.TelegramInfo.DocumentFileID != ""
		if alreadyCached {
			file = telegoutil.FileFromID(picture.TelegramInfo.DocumentFileID)
		} else {
			data, err := storage.GetFile(ctx, picture.StorageInfo.Original)
			if err != nil {
				common.Logger.Errorf("获取文件失败: %s", err)
				utils.ReplyMessage(bot, message, fmt.Sprintf("获取第 %d 张图片失败", i+1))
				return
			}
			file = telegoutil.File(telegoutil.NameReader(bytes.NewReader(data), filepath.Base(picture.StorageInfo.Original.Path)))
		}
		document := telegoutil.Document(message.Chat.ChatID(), file).
			WithReplyParameters(&telego.ReplyParameters{
				MessageID: message.MessageID,
			}).WithCaption(artwork.Title + "_" + strconv.Itoa(i+1))
		if IsChannelAvailable && picture.TelegramInfo.MessageID != 0 {
			document.WithReplyMarkup(telegoutil.InlineKeyboard([]telego.InlineKeyboardButton{
				telegoutil.InlineKeyboardButton("详情").WithURL(utils.GetArtworkPostMessageURL(picture.TelegramInfo.MessageID, ChannelChatID)),
			}))
		} else {
			document.WithReplyMarkup(telegoutil.InlineKeyboard([]telego.InlineKeyboardButton{
				telegoutil.InlineKeyboardButton("详情").WithURL(artwork.SourceURL),
			}))
		}
		documentMessage, err := bot.SendDocument(document)
		if err != nil {
			common.Logger.Errorf("发送文件失败: %s", err)
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
				common.Logger.Warnf("更新图片信息失败: %s", err)
			}
			if alreadyCached {
				time.Sleep(time.Duration(config.Cfg.Telegram.Sleep) * time.Second)
			}
		}
	}
}
