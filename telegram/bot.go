package telegram

import (
	"context"
	"os"
	"time"

	"github.com/krau/ManyACG/common"
	"github.com/krau/ManyACG/config"
	"github.com/krau/ManyACG/service"
	"github.com/krau/ManyACG/telegram/handlers"
	"github.com/krau/ManyACG/telegram/utils"

	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegoapi"
	"github.com/mymmrac/telego/telegohandler"
	"github.com/mymmrac/telego/telegoutil"
)

var (
	Bot                *telego.Bot
	BotUsername        string // 没有 @
	ChannelChatID      telego.ChatID
	GroupChatID        telego.ChatID // 附属群组
	IsChannelAvailable bool          // 是否可以发布到频道
)

var (
	CommonCommands = []telego.BotCommand{
		{
			Command:     "start",
			Description: "开始涩涩",
		},
		{
			Command:     "files",
			Description: "获取作品所有原图文件",
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
			Description: "搜索相似图片",
		},
		{
			Command:     "info",
			Description: "获取作品图片和信息",
		},
		{
			Command:     "hash",
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
		{
			Command:     "hybrid",
			Description: "基于语义与关键字混合搜索作品",
		},
		{
			Command:     "similar",
			Description: "获取与回复的图片相似的作品",
		},
	}

	AdminCommands = []telego.BotCommand{
		{
			Command:     "set_admin",
			Description: "设置管理员",
		},
		{
			Command:     "delete",
			Description: "删除作品",
		},
		{
			Command:     "r18",
			Description: "设置作品 R18",
		},
		{
			Command:     "title",
			Description: "设置作品标题",
		},
		{
			Command:     "tags",
			Description: "设置作品标签(覆盖)",
		},
		{
			Command:     "autotag",
			Description: "自动添加作品标签(AI)",
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
			Command:     "tagalias",
			Description: "为标签添加别名 <原标签名> <别名1> <别名2> ...",
		},
		{
			Command:     "post",
			Description: "发布作品 <url>",
		},
		{
			Command:     "refresh",
			Description: "删除作品缓存 <url>",
		},
		{
			Command:     "recaption",
			Description: "重新生成作品描述 <url>",
		},
	}
)

func InitBot() {
	common.Logger.Info("Initializing bot")
	var err error
	apiUrl := config.Cfg.Telegram.APIURL
	if apiUrl == "" {
		apiUrl = "https://api.telegram.org"
	}
	Bot, err = telego.NewBot(
		config.Cfg.Telegram.Token,
		telego.WithDefaultLogger(false, true),
		telego.WithAPIServer(apiUrl),
		telego.WithAPICaller(&telegoapi.RetryCaller{
			Caller:       telegoapi.DefaultFastHTTPCaller,
			MaxAttempts:  config.Cfg.Telegram.Retry.MaxAttempts,
			ExponentBase: config.Cfg.Telegram.Retry.ExponentBase,
			StartDelay:   time.Duration(config.Cfg.Telegram.Retry.StartDelay),
			MaxDelay:     time.Duration(config.Cfg.Telegram.Retry.MaxDelay),
		}),
	)
	if err != nil {
		common.Logger.Fatalf("Error when creating bot: %s", err)
		os.Exit(1)
	}

	if config.Cfg.Telegram.Username != "" {
		ChannelChatID = telegoutil.Username(config.Cfg.Telegram.Username)
	} else {
		ChannelChatID = telegoutil.ID(config.Cfg.Telegram.ChatID)
	}
	if ChannelChatID.ID == 0 && ChannelChatID.Username == "" {
		if config.Cfg.Telegram.Channel {
			common.Logger.Fatalf("Enabled channel mode but no channel ID or username is provided")
			os.Exit(1)
		}
	} else {
		IsChannelAvailable = config.Cfg.Telegram.Channel
	}

	if config.Cfg.Telegram.GroupID != 0 {
		GroupChatID = telegoutil.ID(config.Cfg.Telegram.GroupID)
	}

	me, err := Bot.GetMe()
	if err != nil {
		common.Logger.Errorf("Error when getting bot info: %s", err)
		os.Exit(1)
	}
	BotUsername = me.Username

	handlers.Init(ChannelChatID, BotUsername)
	utils.Init(ChannelChatID, GroupChatID, BotUsername, NewTelegram())

	Bot.SetMyCommands(&telego.SetMyCommandsParams{
		Commands: CommonCommands,
		Scope:    &telego.BotCommandScopeDefault{Type: telego.ScopeTypeDefault},
	})

	allCommands := append(CommonCommands, AdminCommands...)

	adminUserIDs, err := service.GetAdminUserIDs(context.TODO())
	if err != nil {
		common.Logger.Warnf("Error when getting admin user IDs: %s", err)
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
		common.Logger.Warnf("Error when getting admin group IDs: %s", err)
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
		common.Logger.Errorf("Error when getting bot info: %s", err)
		os.Exit(1)
	}
	common.Logger.Infof("Bot %s is ready", botInfo.Username)

	if service.GetEtcData(context.TODO(), "bot_photo_file_id") != nil && service.GetEtcData(context.TODO(), "bot_photo_bytes") != nil {
		return
	}

	botPhoto, err := Bot.GetUserProfilePhotos(&telego.GetUserProfilePhotosParams{
		UserID: botInfo.ID,
		Limit:  1,
	})
	if err != nil {
		common.Logger.Errorf("Error when getting bot photo: %s", err)
		os.Exit(1)
	}
	if botPhoto.TotalCount == 0 {
		common.Logger.Warn("Please set bot photo")
		os.Exit(1)
	}

	photoSize := botPhoto.Photos[0][len(botPhoto.Photos[0])-1]
	photoFile, err := Bot.GetFile(&telego.GetFileParams{
		FileID: photoSize.FileID,
	})
	if err != nil {
		common.Logger.Errorf("Error when getting bot photo: %s", err)
		os.Exit(1)
	}
	fileBytes, err := telegoutil.DownloadFile(Bot.FileDownloadURL(photoFile.FilePath))
	if err != nil {
		common.Logger.Errorf("Error when downloading bot photo: %s", err)
		os.Exit(1)
	}
	_, err = service.SetEtcData(context.TODO(), "bot_photo_bytes", fileBytes)
	if err != nil {
		common.Logger.Errorf("Error when setting bot photo bytes: %s", err)
		os.Exit(1)
	}
	_, err = service.SetEtcData(context.TODO(), "bot_photo_file_id", photoSize.FileID)
	if err != nil {
		common.Logger.Errorf("Error when setting bot photo file ID: %s", err)
		os.Exit(1)
	}
}

func RunPolling() {
	if Bot == nil {
		InitBot()
	}
	common.Logger.Info("Start polling")
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
		common.Logger.Fatalf("Error when getting updates: %s", err)
		os.Exit(1)
	}

	botHandler, err := telegohandler.NewBotHandler(Bot, updates)
	if err != nil {
		common.Logger.Fatalf("Error when creating bot handler: %s", err)
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
