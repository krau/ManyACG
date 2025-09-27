package telegram

import (
	"context"
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

func InitBot(ctx context.Context) {
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
			StartDelay:   time.Duration(config.Cfg.Telegram.Retry.StartDelay) * time.Second,
			MaxDelay:     time.Duration(config.Cfg.Telegram.Retry.MaxDelay) * time.Second,
			RateLimit:    telegoapi.RetryRateLimitWaitOrAbort,
		}),
	)
	if err != nil {
		common.Logger.Panicf("Error when creating bot: %s", err)
	}

	if config.Cfg.Telegram.Username != "" {
		ChannelChatID = telegoutil.Username(config.Cfg.Telegram.Username)
	} else {
		ChannelChatID = telegoutil.ID(config.Cfg.Telegram.ChatID)
	}
	if ChannelChatID.ID == 0 && ChannelChatID.Username == "" {
		if config.Cfg.Telegram.Channel {
			common.Logger.Panicf("Enabled channel mode but no channel ID or username is provided")
		}
	} else {
		IsChannelAvailable = config.Cfg.Telegram.Channel
	}

	if config.Cfg.Telegram.GroupID != 0 {
		GroupChatID = telegoutil.ID(config.Cfg.Telegram.GroupID)
	}
	me, err := Bot.GetMe(ctx)
	if err != nil {
		common.Logger.Panicf("Error when getting bot info: %s", err)
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
}

func RunPolling(ctx context.Context) {
	if Bot == nil {
		InitBot(ctx)
	}
	common.Logger.Info("Start polling")
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
		common.Logger.Panicf("Error when getting updates: %s", err)
	}

	botHandler, err := telegohandler.NewBotHandler(Bot, updates)
	if err != nil {
		common.Logger.Panicf("Error when creating bot handler: %s", err)
	}
	go func() {
		<-ctx.Done()
		stopCtx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()
		if err := botHandler.StopWithContext(stopCtx); err != nil {
			common.Logger.Warnf("Error when stopping bot handler: %s", err)
		}
		common.Logger.Info("Stopped bot handler")
	}()

	if !config.Cfg.Debug {
		botHandler.Use(telegohandler.PanicRecovery())
	}
	botHandler.Use(messageLogger)

	baseGroup := botHandler.BaseGroup()
	handlers.RegisterHandlers(baseGroup)
	go func() {
		if err := botHandler.Start(); err != nil {
			common.Logger.Panicf("Error when starting bot handler: %s", err)
		}
	}()
	go startService()
}

func startService() {
	go func() {
		for params := range sendArtworkInfoCh {
			if err := utils.SendArtworkInfo(params.Ctx, params.Bot, params.Params); err != nil {
				common.Logger.Errorf("Error when sending artwork info: %s", err)
			}
			time.Sleep(time.Duration(config.Cfg.Telegram.Sleep) * time.Second)
		}
	}()
}
