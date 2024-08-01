package handlers

import (
	"ManyACG/common"
	"ManyACG/service"
	"ManyACG/telegram"
	"ManyACG/types"
	"context"
	"fmt"

	"github.com/mymmrac/telego"
)

func SearchPicture(ctx context.Context, bot *telego.Bot, message telego.Message) {
	hasPermission := CheckPermissionInGroup(ctx, message, types.PermissionSearchPicture)
	if message.ReplyToMessage == nil {
		telegram.ReplyMessage(bot, message, "请使用该命令回复一条图片消息")
		return
	}
	go telegram.ReplyMessage(bot, message, "少女祈祷中...")

	fileBytes, err := telegram.GetMessagePhotoFileBytes(bot, message.ReplyToMessage)
	if err != nil {
		telegram.ReplyMessage(bot, message, "获取图片文件失败: "+err.Error())
		return
	}
	hash, err := common.GetImagePhash(fileBytes)
	if err != nil {
		telegram.ReplyMessage(bot, message, "获取图片哈希失败: "+err.Error())
		return
	}
	pictures, err := service.GetPicturesByHashHammingDistance(ctx, hash, 10)
	if err != nil {
		telegram.ReplyMessage(bot, message, "搜索图片失败: "+err.Error())
		return
	}
	if len(pictures) > 0 {
		text := fmt.Sprintf("找到%d张相似或相同的图片\n\n", len(pictures))
		for _, picture := range pictures {
			artwork, err := service.GetArtworkByMessageID(ctx, picture.TelegramInfo.MessageID)
			if err != nil {
				text += common.EscapeMarkdown(fmt.Sprintf("%s 模糊度: %.2f\n\n", telegram.GetArtworkPostMessageURL(picture.TelegramInfo.MessageID), picture.BlurScore))
			}
			text += fmt.Sprintf("[%s\\_%d](%s)  ",
				common.EscapeMarkdown(artwork.Title),
				picture.Index+1,
				common.EscapeMarkdown(telegram.GetArtworkPostMessageURL(picture.TelegramInfo.MessageID)))
			text += common.EscapeMarkdown(fmt.Sprintf("模糊度: %.2f\n\n", picture.BlurScore))
		}
		telegram.ReplyMessageWithMarkdown(bot, message, text)
		return
	}
	if !hasPermission {
		telegram.ReplyMessage(bot, message, "未在数据库中找到相似图片")
		return
	}
	// TODO: 有权限时使用其他搜索引擎搜图
	telegram.ReplyMessage(bot, message, "未找到相似图片")
}
