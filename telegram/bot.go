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
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	me, err := Bot.GetMe(ctx)
	if err != nil {
		common.Logger.Errorf("Error when getting bot info: %s", err)
		os.Exit(1)
	}
	BotUsername = me.Username

	handlers.Init(ChannelChatID, BotUsername)
	utils.Init(ChannelChatID, GroupChatID, BotUsername, NewTelegram())

	Bot.SetMyCommands(ctx, &telego.SetMyCommandsParams{
		Commands: CommonCommands,
		Scope:    &telego.BotCommandScopeDefault{Type: telego.ScopeTypeDefault},
	})

	allCommands := append(CommonCommands, AdminCommands...)

	adminUserIDs, err := service.GetAdminUserIDs(ctx)
	if err != nil {
		common.Logger.Warnf("Error when getting admin user IDs: %s", err)
		return
	}

	for _, adminID := range adminUserIDs {
		Bot.SetMyCommands(ctx, &telego.SetMyCommandsParams{
			Commands: allCommands,
			Scope: &telego.BotCommandScopeChat{
				Type:   telego.ScopeTypeChat,
				ChatID: telegoutil.ID(adminID),
			},
		})
		if config.Cfg.Telegram.GroupID == 0 {
			continue
		}
		Bot.SetMyCommands(ctx, &telego.SetMyCommandsParams{
			Commands: allCommands,
			Scope: &telego.BotCommandScopeChatMember{
				Type:   telego.ScopeTypeChat,
				ChatID: GroupChatID,
				UserID: adminID,
			},
		})
	}

	adminGroupIDs, err := service.GetAdminGroupIDs(ctx)
	if err != nil {
		common.Logger.Warnf("Error when getting admin group IDs: %s", err)
		return
	}

	for _, adminID := range adminGroupIDs {
		Bot.SetMyCommands(ctx, &telego.SetMyCommandsParams{
			Commands: allCommands,
			Scope: &telego.BotCommandScopeChat{
				Type:   telego.ScopeTypeChat,
				ChatID: telegoutil.ID(adminID),
			},
		})
	}

	botPhoto, err := Bot.GetUserProfilePhotos(ctx, &telego.GetUserProfilePhotosParams{
		UserID: me.ID,
		Limit:  1,
	})
	if err != nil {
		common.Logger.Errorf("Error when getting bot photo: %s", err)
		os.Exit(1)
	}
	photoSize := botPhoto.Photos[0][len(botPhoto.Photos[0])-1]
	photoFile, err := Bot.GetFile(ctx, &telego.GetFileParams{
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
	ctx := context.Background()
	updates, err := Bot.UpdatesViaLongPolling(ctx, &telego.GetUpdatesParams{
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

	if !config.Cfg.Debug {
		botHandler.Use(telegohandler.PanicRecovery())
	}
	botHandler.Use(messageLogger)

	baseGroup := botHandler.BaseGroup()
	handlers.RegisterHandlers(baseGroup)
	if err := botHandler.Start(); err != nil {
		common.Logger.Fatalf("Error when starting bot handler: %s", err)
		os.Exit(1)
	}
}
