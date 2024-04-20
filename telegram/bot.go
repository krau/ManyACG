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
	Bot           *telego.Bot
	BotUsername   string // 没有 @
	ChannelChatID telego.ChatID
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
		ChannelChatID = telegoutil.Username(config.Cfg.Telegram.Username)
	} else {
		ChannelChatID = telegoutil.ID(config.Cfg.Telegram.ChatID)
	}

	me, err := Bot.GetMe()
	if err != nil {
		Logger.Errorf("Error when getting bot info: %s", err)
		os.Exit(1)
	}
	BotUsername = me.Username

	Bot.SetMyCommands(&telego.SetMyCommandsParams{
		Commands: []telego.BotCommand{
			{
				Command:     "start",
				Description: "开始涩涩",
			},
			{
				Command:     "file",
				Description: "获取原图文件",
			},
			{
				Command:     "setu",
				Description: "来点涩图",
			},
			{
				Command:     "random",
				Description: "随机1张全年龄图片",
			},
		},
	})

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
	botHandler.HandleMessageCtx(randomPicture, telegohandler.Or(telegohandler.CommandEqual("setu"), telegohandler.CommandEqual("random")))

	botHandler.Start()
}
