package utils

import (
	"ManyACG/config"

	"github.com/mymmrac/telego"
)

var (
	ChannelChatID      telego.ChatID
	GroupChatID        telego.ChatID
	BotUsername        string
	IsChannelAvailable bool
)

func Init(channelChatID, groupChatID telego.ChatID, botUsername string) {
	ChannelChatID = channelChatID
	GroupChatID = groupChatID
	BotUsername = botUsername
	IsChannelAvailable = (ChannelChatID.ID != 0 || ChannelChatID.Username != "") && config.Cfg.Telegram.Channel
}
