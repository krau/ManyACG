package telegram

import (
	"ManyACG/config"
	"os"

	. "ManyACG/logger"

	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegoutil"
)

var (
	Bot           *telego.Bot
	BotUsername   string // 没有 @
	ChannelChatID telego.ChatID
	GroupChatID   telego.ChatID // 附属群组
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

	if config.Cfg.Telegram.GroupID != 0 {
		GroupChatID = telegoutil.ID(config.Cfg.Telegram.GroupID)
	}

	me, err := Bot.GetMe()
	if err != nil {
		Logger.Errorf("Error when getting bot info: %s", err)
		os.Exit(1)
	}
	BotUsername = me.Username

	commonCommands := []telego.BotCommand{
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
		{
			Command:     "search",
			Description: "搜索图片",
		},
		{
			Command:     "help",
			Description: "食用指南",
		},
	}

	Bot.SetMyCommands(&telego.SetMyCommandsParams{
		Commands: commonCommands,
		Scope:    &telego.BotCommandScopeDefault{Type: telego.ScopeTypeDefault},
	})

	adminCommands := []telego.BotCommand{
		{
			Command:     "set_admin",
			Description: "设置管理员",
		},
		{
			Command:     "del",
			Description: "删除图片",
		},
		{
			Command:     "delete",
			Description: "删除图片对应的作品",
		},
		{
			Command:     "r18",
			Description: "设置作品 R18",
		},
		{
			Command:     "tags",
			Description: "设置作品标签(覆盖)",
		},
		{
			Command:     "addtags",
			Description: "添加作品标签",
		},
		{
			Command:     "deltags",
			Description: "删除作品标签",
		},
		{
			Command:     "fetch",
			Description: "开始一次拉取",
		},
		{
			Command:     "process_pictures",
			Description: "处理无哈希的图片",
		},
	}

	adminCommands = append(adminCommands, commonCommands...)

	for _, adminID := range config.Cfg.Telegram.Admins {
		Bot.SetMyCommands(&telego.SetMyCommandsParams{
			Commands: adminCommands,
			Scope: &telego.BotCommandScopeChat{
				Type:   telego.ScopeTypeChat,
				ChatID: telegoutil.ID(adminID),
			},
		})
		if config.Cfg.Telegram.GroupID == 0 {
			continue
		}
		Bot.SetMyCommands(&telego.SetMyCommandsParams{
			Commands: adminCommands,
			Scope: &telego.BotCommandScopeChatMember{
				Type:   telego.ScopeTypeChat,
				ChatID: GroupChatID,
				UserID: adminID,
			},
		})
	}
}
