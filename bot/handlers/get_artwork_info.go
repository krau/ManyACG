package handlers

import (
	. "ManyACG/logger"
	"ManyACG/telegram"
	"ManyACG/types"
	"context"
	"time"

	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegoutil"
)

func GetArtworkInfo(ctx context.Context, bot *telego.Bot, message telego.Message) {
	hasPermission := CheckPermissionInGroup(ctx, message, types.PermissionGetArtworkInfo)
	sourceURL := telegram.FindSourceURLForMessage(&message)
	var waitMessageID int
	if hasPermission {
		go func() {
			msg, err := telegram.ReplyMessage(bot, message, "正在获取作品信息...")
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

	err := telegram.SendArtworkInfo(ctx, bot, sourceURL, false, &chatID, hasPermission, "", false, &telego.ReplyParameters{
		MessageID: message.MessageID,
	})
	if err != nil {
		Logger.Error(err)
		telegram.ReplyMessage(bot, message, err.Error())
	}
}
