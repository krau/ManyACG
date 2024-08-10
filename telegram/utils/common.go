package utils

import "github.com/mymmrac/telego"

var (
	ChannelChatID telego.ChatID
	GroupChatID   telego.ChatID
	BotUsername   string
)

func Init(channelChatID, groupChatID telego.ChatID, botUsername string) {
	ChannelChatID = channelChatID
	GroupChatID = groupChatID
	BotUsername = botUsername
}
