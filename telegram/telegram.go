package telegram

import (
	"ManyACG-Bot/config"
	"os"

	. "ManyACG-Bot/logger"

	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegoutil"
)

var (
	Bot           *telego.Bot
	BotUsername   string // 没有 @
	ChannelChatID telego.ChatID
)

func InitBot() {
	var err error
	Bot, err = telego.NewBot(
		config.Cfg.Telegram.Token,
		telego.WithDefaultLogger(false, true),
		telego.WithAPIServer(config.Cfg.Telegram.APIURL),
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
				Description: "来点涩图 <tag1> <tag2> ...",
			},
			{
				Command:     "random",
				Description: "随机1张全年龄图片 <tag1> <tag2> ...",
			},
			{
				Command:     "search",
				Description: "搜索图片",
			},
			{
				Command:     "help",
				Description: "食用指南",
			},
		},
	})
}
