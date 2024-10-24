package handlers

import (
	"context"
	"fmt"

	"github.com/krau/ManyACG/common"
	"github.com/krau/ManyACG/config"
	. "github.com/krau/ManyACG/logger"
	"github.com/krau/ManyACG/service"
	"github.com/krau/ManyACG/telegram/utils"
	"github.com/krau/ManyACG/types"

	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegoutil"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func SearchPicture(ctx context.Context, bot *telego.Bot, message telego.Message) {
	hasPermission := CheckPermissionInGroup(ctx, message, types.PermissionSearchPicture)
	if message.ReplyToMessage == nil {
		utils.ReplyMessage(bot, message, "请使用该命令回复一条图片消息")
		return
	}
	go utils.ReplyMessage(bot, message, "少女祈祷中...")

	fileBytes, err := utils.GetMessagePhotoFileBytes(bot, message.ReplyToMessage)
	if err != nil {
		utils.ReplyMessage(bot, message, "获取图片文件失败: "+err.Error())
		return
	}
	text, _, err := getSearchResult(ctx, hasPermission, fileBytes)
	if err != nil {
		utils.ReplyMessage(bot, message, err.Error())
		return
	}
	utils.ReplyMessageWithMarkdown(bot, message, text)
}

func SearchPictureCallbackQuery(ctx context.Context, bot *telego.Bot, query telego.CallbackQuery) {
	if !query.Message.IsAccessible() {
		return
	}
	message := query.Message.(*telego.Message)
	fileBytes, err := utils.GetMessagePhotoFileBytes(bot, message)
	if err != nil {
		bot.AnswerCallbackQuery(telegoutil.CallbackQuery(query.ID).WithText("获取图片文件失败: " + err.Error()).WithShowAlert().WithCacheTime(5))
		return
	}
	text, _, err := getSearchResult(ctx, true, fileBytes)
	if err != nil {
		bot.AnswerCallbackQuery(telegoutil.CallbackQuery(query.ID).WithText(err.Error()).WithShowAlert().WithCacheTime(5))
		return
	}
	go bot.AnswerCallbackQuery(telegoutil.CallbackQuery(query.ID).WithText(text).WithCacheTime(5))
	utils.ReplyMessageWithMarkdown(bot, *message, text)
}

func getSearchResult(ctx context.Context, hasPermission bool, fileBytes []byte) (string, bool, error) {
	hash, err := common.GetImagePhash(fileBytes)
	if err != nil {
		return "", false, fmt.Errorf("获取图片哈希失败: %w", err)
	}
	pictures, err := service.GetPicturesByHashHammingDistance(ctx, hash, 10)
	if err != nil {
		return "", false, fmt.Errorf("搜索图片失败: %w", err)
	}
	channelMessageAvailable := ChannelChatID.ID != 0 || ChannelChatID.Username != ""
	if len(pictures) > 0 {
		text := fmt.Sprintf("找到%d张相似的图片\n\n", len(pictures))
		for _, picture := range pictures {
			artworkObjectID, err := primitive.ObjectIDFromHex(picture.ArtworkID)
			if err != nil {
				Logger.Errorf("无效的ObjectID: %s", picture.ID)
				continue
			}
			artwork, err := service.GetArtworkByID(ctx, artworkObjectID)
			if err != nil {
				Logger.Errorf("获取作品信息失败: %s", err)
				continue
			}
			text += fmt.Sprintf("[%s\\_%d](%s)\n",
				common.EscapeMarkdown(artwork.Title),
				picture.Index+1,
				common.EscapeMarkdown(artwork.SourceURL),
			)
			if channelMessageAvailable && picture.TelegramInfo != nil && picture.TelegramInfo.MessageID != 0 {
				text += fmt.Sprintf("[频道消息](%s)\n", utils.GetArtworkPostMessageURL(picture.TelegramInfo.MessageID, ChannelChatID))
			}
			text += fmt.Sprintf("[ManyACG](%s)\n", config.Cfg.API.SiteURL+"/artwork/"+artwork.ID)
			text += common.EscapeMarkdown(fmt.Sprintf("模糊度: %.2f\n\n", picture.BlurScore))
		}
		return text, true, nil
	}
	if !hasPermission {
		return "未在数据库中找到相似图片", false, nil
	}
	// TODO: 有权限时使用其他搜索引擎搜图
	return "未找到相似图片", false, nil
}
