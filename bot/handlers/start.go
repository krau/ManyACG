package handlers

import (
	"ManyACG/telegram"
	"context"
	"strconv"
	"strings"

	. "ManyACG/logger"

	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegoutil"
)

func Start(ctx context.Context, bot *telego.Bot, message telego.Message) {
	_, _, args := telegoutil.ParseCommand(message.Text)
	if len(args) > 0 {
		Logger.Debugf("start: args=%v", args)
		if strings.HasPrefix(args[0], "file_") {
			messageIDStr := args[0][5:]
			messageID, err := strconv.Atoi(messageIDStr)
			if err != nil {
				telegram.ReplyMessage(bot, message, "获取失败: "+err.Error())
				return
			}
			_, err = telegram.SendPictureFileByMessageID(ctx, bot, message, messageID)
			if err != nil {
				telegram.ReplyMessage(bot, message, "获取失败: "+err.Error())
				return
			}
		}
		return
	}
	Help(ctx, bot, message)
}
