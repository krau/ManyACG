package bot

import (
	"github.com/mymmrac/telego"
)

func sourceUrlMatches(update telego.Update) bool {
	return MatchSourceURLForMessage(update.Message) != ""
}
