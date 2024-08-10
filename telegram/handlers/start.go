package handlers

import (
	"ManyACG/telegram/utils"
	"context"
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
			pictureID := args[0][5:]
			_, err := utils.SendPictureFileByID(ctx, bot, message, ChannelChatID, pictureID)
			if err != nil {
				utils.ReplyMessage(bot, message, "获取失败: "+err.Error())
				return
			}
		}
		return
	}
	Help(ctx, bot, message)
}
