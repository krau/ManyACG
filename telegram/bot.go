package telegram

import (
	"ManyACG-Bot/config"
	"os"

	. "ManyACG-Bot/logger"

	"github.com/mymmrac/telego"
	tu "github.com/mymmrac/telego/telegoutil"
)

var (
	Bot    *telego.Bot
	ChatID telego.ChatID
)

func init() {
	var err error
	Bot, err = telego.NewBot(
		config.Cfg.Telegram.Token,
		telego.WithDefaultLogger(false, true),
		// telego.WithAPICaller(
		// 	&telegoapi.RetryCaller{
		// 		Caller:       telegoapi.DefaultFastHTTPCaller,
		// 		MaxAttempts:  3,
		// 		ExponentBase: 2,
		// 		StartDelay:   10,
		// 	},
		// ),
	)
	if err != nil {
		Logger.Fatalf("Error when creating bot: %s", err)
		os.Exit(1)
	}
	if config.Cfg.Telegram.Username != "" {
		ChatID = tu.Username(config.Cfg.Telegram.Username)
	} else {
		ChatID = tu.ID(config.Cfg.Telegram.ChatID)
	}
}
