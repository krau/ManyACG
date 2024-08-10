package telegram

import (
	"ManyACG/config"
	"ManyACG/service"
	"ManyACG/telegram/handlers"
	"ManyACG/telegram/utils"
	"context"
	"os"

	. "ManyACG/logger"

	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegohandler"
	"github.com/mymmrac/telego/telegoutil"
)

var (
	Bot           *telego.Bot
	BotUsername   string // 没有 @
	ChannelChatID telego.ChatID
	GroupChatID   telego.ChatID // 附属群组
)

var (
	CommonCommands = []telego.BotCommand{
		{
			Command:     "start",
			Description: "开始涩涩",
		},
		{
			Command:     "file",
			Description: "获取原图文件",
		},
		{
			Command:     "files",
			Description: "获取作品所有图片文件",
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
			Command:     "info",
			Description: "计算图片信息",
		},
		{
			Command:     "stats",
			Description: "统计数据",
		},
		{
			Command:     "help",
			Description: "食用指南",
		},
	}

	AdminCommands = []telego.BotCommand{
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
			Command:     "post",
			Description: "发布作品 <url>",
		},
		{
			Command:     "process_pictures",
			Description: "处理无哈希的图片",
		},
	}
)

func InitBot() {
	Logger.Info("Initializing bot")
	var err error
	apiUrl := config.Cfg.Telegram.APIURL
	if apiUrl == "" {
		apiUrl = "https://api.telegram.org"
	}
	Bot, err = telego.NewBot(
		config.Cfg.Telegram.Token,
		telego.WithDefaultLogger(false, true),
		telego.WithAPIServer(apiUrl),
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

	handlers.Init(ChannelChatID, BotUsername)
	utils.Init(ChannelChatID, GroupChatID, BotUsername)

	Bot.SetMyCommands(&telego.SetMyCommandsParams{
		Commands: CommonCommands,
		Scope:    &telego.BotCommandScopeDefault{Type: telego.ScopeTypeDefault},
	})

	allCommands := append(CommonCommands, AdminCommands...)

	adminUserIDs, err := service.GetAdminUserIDs(context.TODO())
	if err != nil {
		Logger.Warnf("Error when getting admin user IDs: %s", err)
		return
	}

	for _, adminID := range adminUserIDs {
		Bot.SetMyCommands(&telego.SetMyCommandsParams{
			Commands: allCommands,
			Scope: &telego.BotCommandScopeChat{
				Type:   telego.ScopeTypeChat,
				ChatID: telegoutil.ID(adminID),
			},
		})
		if config.Cfg.Telegram.GroupID == 0 {
			continue
		}
		Bot.SetMyCommands(&telego.SetMyCommandsParams{
			Commands: allCommands,
			Scope: &telego.BotCommandScopeChatMember{
				Type:   telego.ScopeTypeChat,
				ChatID: GroupChatID,
				UserID: adminID,
			},
		})
	}

	adminGroupIDs, err := service.GetAdminGroupIDs(context.TODO())
	if err != nil {
		Logger.Warnf("Error when getting admin group IDs: %s", err)
		return
	}

	for _, adminID := range adminGroupIDs {
		Bot.SetMyCommands(&telego.SetMyCommandsParams{
			Commands: allCommands,
			Scope: &telego.BotCommandScopeChat{
				Type:   telego.ScopeTypeChat,
				ChatID: telegoutil.ID(adminID),
			},
		})
	}
	botInfo, err := Bot.GetMe()
	if err != nil {
		Logger.Errorf("Error when getting bot info: %s", err)
		os.Exit(1)
	}
	Logger.Infof("Bot %s is ready", botInfo.Username)

	if service.GetEtcData(context.TODO(), "bot_photo_file_id") != nil && service.GetEtcData(context.TODO(), "bot_photo_bytes") != nil {
		return
	}

	botPhoto, err := Bot.GetUserProfilePhotos(&telego.GetUserProfilePhotosParams{
		UserID: botInfo.ID,
		Limit:  1,
	})
	if err != nil {
		Logger.Errorf("Error when getting bot photo: %s", err)
		os.Exit(1)
	}
	if botPhoto.TotalCount == 0 {
		Logger.Warn("Please set bot photo")
		os.Exit(1)
	}

	photoSize := botPhoto.Photos[0][len(botPhoto.Photos[0])-1]
	photoFile, err := Bot.GetFile(&telego.GetFileParams{
		FileID: photoSize.FileID,
	})
	if err != nil {
		Logger.Errorf("Error when getting bot photo: %s", err)
		os.Exit(1)
	}
	fileBytes, err := telegoutil.DownloadFile(Bot.FileDownloadURL(photoFile.FilePath))
	if err != nil {
		Logger.Errorf("Error when downloading bot photo: %s", err)
		os.Exit(1)
	}
	_, err = service.SetEtcData(context.TODO(), "bot_photo_bytes", fileBytes)
	if err != nil {
		Logger.Errorf("Error when setting bot photo bytes: %s", err)
		os.Exit(1)
	}
	_, err = service.SetEtcData(context.TODO(), "bot_photo_file_id", photoSize.FileID)
	if err != nil {
		Logger.Errorf("Error when setting bot photo file ID: %s", err)
		os.Exit(1)
	}
}

func RunPolling() {
	if Bot == nil {
		InitBot()
	}
	Logger.Info("Start polling")
	updates, err := Bot.UpdatesViaLongPolling(&telego.GetUpdatesParams{
		Offset: -1,
		AllowedUpdates: []string{
			telego.MessageUpdates,
			telego.ChannelPostUpdates,
			telego.CallbackQueryUpdates,
			telego.InlineQueryUpdates,
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

	if !config.Cfg.Debug {
		botHandler.Use(telegohandler.PanicRecovery())
	}
	botHandler.Use(messageLogger)
	baseGroup := botHandler.BaseGroup()
	handlers.RegisterHandlers(baseGroup)
	botHandler.Start()
}
