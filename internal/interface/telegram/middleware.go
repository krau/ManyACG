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
			log.Debugf("[%s](%d) [%s](%s)", chat.Title, chat.ID, senderChat.Title, senderChat.Username)
		} else {
			log.Debugf("[%s](%d) [%s](%d)", chat.Title, chat.ID, user.FirstName+user.LastName, user.ID)
		}
	}
	return ctx.Next(update)
}
