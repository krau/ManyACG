package bot

import (
	. "ManyACG-Bot/logger"

	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegohandler"
)

func messageLogger(bot *telego.Bot, update telego.Update, next telegohandler.Handler) {
	if update.Message != nil {
		chat := update.Message.Chat
		user := update.Message.From
		senderChat := update.Message.SenderChat
		if senderChat != nil {
			Logger.Tracef("[%s](%d) [%s](%d): %s", chat.Title, chat.ID, senderChat.Title, senderChat.Username, update.Message.Text)
		} else {
			Logger.Tracef("[%s](%d) [%s](%d): %s", chat.Title, chat.ID, user.FirstName+user.LastName, user.ID, update.Message.Text)
		}
	}

	next(bot, update)
}
