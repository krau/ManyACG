package bot

import "github.com/mymmrac/telego"

func onlyPrivate(update telego.Update) bool {
	return update.Message != nil && update.Message.Chat.Type == telego.ChatTypePrivate
}
