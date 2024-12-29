package handlers

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"path"
	"path/filepath"
	"strconv"
	"strings"
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
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func GetPictureFile(ctx context.Context, bot *telego.Bot, message telego.Message) {
	var sourceURL string
	if message.ReplyToMessage != nil {
		sourceURL = utils.FindSourceURLForMessage(message.ReplyToMessage)
	} else {
		sourceURL = sources.FindSourceURL(message.Text)
	}
	cmd, _, args := telegoutil.ParseCommand(message.Text)
	multiple := cmd == "files"
	if sourceURL == "" {
		getPictureByHash := func() *types.Picture {
			if message.ReplyToMessage == nil {
				return nil
			}
			file, err := utils.GetMessagePhotoFile(bot, message.ReplyToMessage)
			if err != nil {
				return nil
			}
			hash, err := common.GetImagePhashFromReader(bytes.NewReader(file))
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
		if !multiple {
			_, err := utils.SendPictureFileByID(ctx, bot, message, ChannelChatID, picture.ID)
			if err != nil {
				utils.ReplyMessage(bot, message, "文件发送失败: "+err.Error())
				return
			}
		} else {
			artworkID, err := primitive.ObjectIDFromHex(picture.ArtworkID)
			if err != nil {
				utils.ReplyMessage(bot, message, "无效的作品 ID")
				return
			}
			artwork, err := service.GetArtworkByID(ctx, artworkID)
			if err != nil {
				utils.ReplyMessage(bot, message, "获取作品信息失败")
				return
			}
			getArtworkFiles(ctx, bot, message, artwork)
		}
		return
	}

	artwork, err := service.GetArtworkByURL(ctx, sourceURL)
	if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
		common.Logger.Errorf("获取作品信息失败: %s", err)
		utils.ReplyMessage(bot, message, "获取作品信息失败")
		return
	}
	if multiple {
		var err error
		if artwork == nil {
			artwork, err = service.GetArtworkByURLWithCacheFetch(ctx, sourceURL)
			if err != nil {
				common.Logger.Errorf("获取作品信息失败: %s", err)
				utils.ReplyMessage(bot, message, "获取作品信息失败")
				return
			}
		}
		getArtworkFiles(ctx, bot, message, artwork)
		return
	}
	if artwork == nil {
		utils.ReplyMessage(bot, message, "未找到作品信息")
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
	for i, picture := range artwork.Pictures {
		var file telego.InputFile
		alreadyCached := picture.TelegramInfo != nil && picture.TelegramInfo.DocumentFileID != ""
		if alreadyCached {
			file = telegoutil.FileFromID(picture.TelegramInfo.DocumentFileID)
		} else if picture.StorageInfo != nil && picture.StorageInfo.Original != nil {
			data, err := storage.GetFile(ctx, picture.StorageInfo.Original)
			if err != nil {
				data, err = common.DownloadWithCache(ctx, picture.Original, nil)
				if err != nil {
					common.Logger.Errorf("获取文件失败: %s", err)
					utils.ReplyMessage(bot, message, fmt.Sprintf("获取第 %d 张图片失败", i+1))
					return
				}
			}
			filename := func() string {
				if picture.StorageInfo.Original != nil && picture.StorageInfo.Original.Path != "" {
					return filepath.Base(picture.StorageInfo.Original.Path)
				}
				return path.Base(strings.Split(picture.Original, "?")[0])
			}()
			file = telegoutil.File(telegoutil.NameReader(bytes.NewReader(data), filename))
		} else {
			data, err := common.DownloadWithCache(ctx, picture.Original, nil)
			if err != nil {
				common.Logger.Errorf("获取文件失败: %s", err)
				utils.ReplyMessage(bot, message, fmt.Sprintf("获取第 %d 张图片失败", i+1))
				return
			}
			file = telegoutil.File(telegoutil.NameReader(bytes.NewReader(data), path.Base(strings.Split(picture.Original, "?")[0])))
		}
		document := telegoutil.Document(message.Chat.ChatID(), file).
			WithReplyParameters(&telego.ReplyParameters{
				MessageID: message.MessageID,
			}).WithCaption(artwork.Title + "_" + strconv.Itoa(i+1)).WithDisableContentTypeDetection()
		if IsChannelAvailable && picture.TelegramInfo != nil && picture.TelegramInfo.MessageID != 0 {
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
			if picture.TelegramInfo == nil {
				picture.TelegramInfo = &types.TelegramInfo{}
			}
			picture.TelegramInfo.DocumentFileID = documentMessage.Document.FileID
			if err := service.UpdatePictureTelegramInfo(ctx, picture, picture.TelegramInfo); err != nil {
				common.Logger.Warnf("更新图片信息失败: %s", err)
			}
		}
		time.Sleep(time.Duration(config.Cfg.Telegram.Sleep) * time.Second)
	}
}
