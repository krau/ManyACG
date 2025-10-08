package telegram

import (
	"github.com/krau/ManyACG/pkg/log"
	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegohandler"
)

func messageLogger(ctx *telegohandler.Context, update telego.Update) error {
	if update.Message != nil {
		chat := update.Message.Chat
		user := update.Message.From
		senderChat := update.Message.SenderChat
		if senderChat != nil {
			log.Debug("received message", "chat_title", chat.Title, "chat_id", chat.ID, "sender", senderChat.Title, "sender_username", senderChat.Username)
		} else {
			log.Debug("received message", "chat_title", chat.Title, "chat_id", chat.ID, "sender", user.FirstName+user.LastName, "sender_username", user.ID)
		}
	}
	return ctx.Next(update)
}
