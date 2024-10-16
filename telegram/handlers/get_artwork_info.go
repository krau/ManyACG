package handlers

import (
	"context"
	"time"

	. "github.com/krau/ManyACG/logger"
	"github.com/krau/ManyACG/service"
	"github.com/krau/ManyACG/telegram/utils"
	"github.com/krau/ManyACG/types"

	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegoutil"
)

func GetArtworkInfo(ctx context.Context, bot *telego.Bot, message telego.Message) {
	hasPermission := CheckPermissionInGroup(ctx, message, types.PermissionGetArtworkInfo)
	sourceURL := utils.FindSourceURLForMessage(&message)
	var waitMessageID int
	if hasPermission {
		go func() {
			msg, err := utils.ReplyMessage(bot, message, "正在获取作品信息...")
			if err != nil {
				Logger.Warnf("发送消息失败: %s", err)
				return
			}
			waitMessageID = msg.MessageID
		}()
	}
	defer func() {
		time.Sleep(1 * time.Second)
		if waitMessageID != 0 {
			bot.DeleteMessage(telegoutil.Delete(message.Chat.ChatID(), waitMessageID))
		}
	}()
	chatID := message.Chat.ChatID()

	err := utils.SendArtworkInfo(ctx, bot, &utils.SendArtworkInfoParams{
		ChatID:        &chatID,
		SourceURL:     sourceURL,
		AppendCaption: "",
		Verify:        false,
		IgnoreDeleted: false,
		HasPermission: hasPermission,
		ReplyParams: &telego.ReplyParameters{
			MessageID: message.MessageID,
		},
	})
	if err != nil {
		Logger.Error(err)
		utils.ReplyMessage(bot, message, err.Error())
	}
}

func GetArtworkInfoCommand(ctx context.Context, bot *telego.Bot, message telego.Message) {
	sourceURL := utils.FindSourceURLForMessage(&message)
	if sourceURL == "" {
		sourceURL = utils.FindSourceURLForMessage(message.ReplyToMessage)
	}
	if sourceURL == "" {
		utils.ReplyMessage(bot, message, "命令参数或回复的消息中没有找到支持的链接")
		return
	}
	var waitMessageID int
	go func() {
		msg, err := utils.ReplyMessage(bot, message, "正在获取作品信息...")
		if err != nil {
			Logger.Warnf("发送消息失败: %s", err)
			return
		}
		waitMessageID = msg.MessageID
	}()
	defer func() {
		time.Sleep(1 * time.Second)
		if waitMessageID != 0 {
			bot.DeleteMessage(telegoutil.Delete(message.Chat.ChatID(), waitMessageID))
		}
	}()
	artwork, err := service.GetArtworkByURLWithCacheFetch(ctx, sourceURL)
	if err != nil {
		Logger.Error(err)
		utils.ReplyMessage(bot, message, err.Error())
		return
	}
	messages, err := utils.SendArtworkMediaGroup(ctx, bot, message.Chat.ChatID(), artwork)
	if err != nil {
		Logger.Error(err)
		utils.ReplyMessage(bot, message, "发送作品信息失败")
	}

	cachedArtwork, err := service.GetCachedArtworkByURL(ctx, sourceURL)
	if err != nil {
		Logger.Warnf("获取缓存作品信息失败: %s", err)
		return
	}

	for i, picture := range cachedArtwork.Artwork.Pictures {
		if picture.TelegramInfo == nil {
			picture.TelegramInfo = &types.TelegramInfo{}
		}
		if i < len(messages) {
			if messages[i].Photo != nil {
				picture.TelegramInfo.PhotoFileID = messages[i].Photo[len(messages[i].Photo)-1].FileID
			}
		}
	}
	if err := service.UpdateCachedArtwork(ctx, cachedArtwork); err != nil {
		Logger.Warnf("更新缓存作品信息失败: %s", err)
	}
}
