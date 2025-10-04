package filter

import (
	"context"
	"strings"

	"github.com/krau/ManyACG/internal/interface/telegram/metautil"
	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegoutil"
)

// 如果在群里使用命令且包含 @botusername, 则判断是否是本机器人, 不是则返回 false
//
// 其他情况下默认返回 true
func CommandToMe(ctx context.Context, update telego.Update) bool {
	meta := metautil.FromContext(ctx)
	if update.Message.Chat.Type != telego.ChatTypePrivate {
		_, botUsername, _ := telegoutil.ParseCommand(update.Message.Text)
		if botUsername == "" {
			return true
		}
		return strings.TrimPrefix(botUsername, "@") == meta.BotUsername
	}
	return true
}
