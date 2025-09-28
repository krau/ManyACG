package utils

import (
	"github.com/krau/ManyACG/internal/infra/config"
	"github.com/krau/ManyACG/types"

	"github.com/mymmrac/telego"
)

var (
	ChannelChatID      telego.ChatID
	GroupChatID        telego.ChatID
	BotUsername        string
	IsChannelAvailable bool
	TgService          types.TelegramService
)

func Init(channelChatID, groupChatID telego.ChatID, botUsername string, tgService types.TelegramService) {
	ChannelChatID = channelChatID
	GroupChatID = groupChatID
	BotUsername = botUsername
	IsChannelAvailable = (ChannelChatID.ID != 0 || ChannelChatID.Username != "") && config.Get().Telegram.Channel
	TgService = tgService
}
