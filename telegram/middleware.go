package telegram

import (
	"github.com/krau/ManyACG/common"
	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegohandler"
)

func messageLogger(bot *telego.Bot, update telego.Update, next telegohandler.Handler) {
	if update.Message != nil {
		chat := update.Message.Chat
		user := update.Message.From
		senderChat := update.Message.SenderChat
		if senderChat != nil {
			common.Logger.Tracef("[%s](%d) [%s](%s)", chat.Title, chat.ID, senderChat.Title, senderChat.Username)
		} else {
			common.Logger.Tracef("[%s](%d) [%s](%d)", chat.Title, chat.ID, user.FirstName+user.LastName, user.ID)
		}
	}

	next(bot, update)
}
