package bot

import (
	. "ManyACG-Bot/logger"
	"ManyACG-Bot/service"
	"ManyACG-Bot/telegram"

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

func adminCheck(bot *telego.Bot, update telego.Update, next telegohandler.Handler) {
	if update.Message == nil && update.CallbackQuery == nil {
		return
	}
	var userID int64
	if update.Message != nil {
		userID = update.Message.From.ID
	}
	if update.CallbackQuery != nil {
		userID = update.CallbackQuery.From.ID
	}
	isAdmin, err := service.IsAdmin(update.Context(), userID)
	if !isAdmin {
		Logger.Debugf("User %d is not admin: %s", userID, err)
		if update.CallbackQuery != nil {
			bot.AnswerCallbackQuery(&telego.AnswerCallbackQueryParams{
				CallbackQueryID: update.CallbackQuery.ID,
				Text:            "你没有权限哦",
				ShowAlert:       true,
				CacheTime:       60,
			})
			return
		}
		if update.Message != nil {
			telegram.ReplyMessage(bot, *update.Message, "你没有权限哦")
			return
		}
		return
	}
	next(bot, update)
}
