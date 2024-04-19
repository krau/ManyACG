package telegram

import (
	"ManyACG-Bot/config"
	"os"

	. "ManyACG-Bot/logger"

	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegohandler"
	"github.com/mymmrac/telego/telegoutil"
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
	)
	if err != nil {
		Logger.Fatalf("Error when creating bot: %s", err)
		os.Exit(1)
	}
	if config.Cfg.Telegram.Username != "" {
		ChatID = telegoutil.Username(config.Cfg.Telegram.Username)
	} else {
		ChatID = telegoutil.ID(config.Cfg.Telegram.ChatID)
	}

	go RunPolling()
}

func RunPolling() {
	Logger.Info("Start polling")
	updates, err := Bot.UpdatesViaLongPolling(&telego.GetUpdatesParams{
		Offset: -1,
		AllowedUpdates: []string{
			telego.MessageUpdates,
			telego.ChannelPostUpdates,
			telego.CallbackQueryUpdates,
		},
	})
	if err != nil {
		Logger.Fatalf("Error when getting updates: %s", err)
		os.Exit(1)
	}

	botHandler, err := telegohandler.NewBotHandler(Bot, updates)
	if err != nil {
		Logger.Fatalf("Error when creating bot handler: %s", err)
		os.Exit(1)
	}
	defer botHandler.Stop()
	defer Bot.StopLongPolling()

	botHandler.Use(messageLogger)

	botHandler.HandleMessageCtx(start, telegohandler.CommandEqual("start"))
	botHandler.HandleMessageCtx(getPictureFile, telegohandler.CommandEqual("file"))

	botHandler.Start()
}
