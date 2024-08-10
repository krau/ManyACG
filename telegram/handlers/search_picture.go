package handlers

import (
	"ManyACG/common"
	"ManyACG/service"
	"ManyACG/telegram/utils"
	"ManyACG/types"
	"context"
	"fmt"

	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegoutil"
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
	text, err := getSearchResult(ctx, hasPermission, fileBytes)
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
	go bot.AnswerCallbackQuery(telegoutil.CallbackQuery(query.ID).WithText("少女祈祷中...").WithCacheTime(5))
	fileBytes, err := utils.GetMessagePhotoFileBytes(bot, message)
	if err != nil {
		bot.AnswerCallbackQuery(telegoutil.CallbackQuery(query.ID).WithText("获取图片文件失败: " + err.Error()).WithShowAlert().WithCacheTime(5))
		return
	}
	text, err := getSearchResult(ctx, true, fileBytes)
	if err != nil {
		bot.AnswerCallbackQuery(telegoutil.CallbackQuery(query.ID).WithText(err.Error()).WithShowAlert().WithCacheTime(5))
		return
	}
	utils.ReplyMessageWithMarkdown(bot, *message, text)

}

func getSearchResult(ctx context.Context, hasPermission bool, fileBytes []byte) (string, error) {
	hash, err := common.GetImagePhash(fileBytes)
	if err != nil {
		return "", fmt.Errorf("获取图片哈希失败: %w", err)
	}
	pictures, err := service.GetPicturesByHashHammingDistance(ctx, hash, 10)
	if err != nil {
		return "", fmt.Errorf("搜索图片失败: %w", err)
	}
	if len(pictures) > 0 {
		text := fmt.Sprintf("找到%d张相似的图片\n\n", len(pictures))
		for _, picture := range pictures {
			artwork, err := service.GetArtworkByMessageID(ctx, picture.TelegramInfo.MessageID)
			if err != nil {
				text += common.EscapeMarkdown(fmt.Sprintf("%s 模糊度: %.2f\n\n", utils.GetArtworkPostMessageURL(picture.TelegramInfo.MessageID, ChannelChatID), picture.BlurScore))
				continue
			}
			text += fmt.Sprintf("[%s\\_%d](%s)\n[%s](%s)\n",
				common.EscapeMarkdown(artwork.Title),
				picture.Index+1,
				common.EscapeMarkdown(artwork.SourceURL),
				"\\-\\>频道消息",
				utils.GetArtworkPostMessageURL(picture.TelegramInfo.MessageID, ChannelChatID),
			)
			text += common.EscapeMarkdown(fmt.Sprintf("模糊度: %.2f\n\n", picture.BlurScore))
		}
		return text, nil
	}
	if !hasPermission {
		return "未在数据库中找到相似图片", nil
	}
	// TODO: 有权限时使用其他搜索引擎搜图
	return "未找到相似图片", nil
}
