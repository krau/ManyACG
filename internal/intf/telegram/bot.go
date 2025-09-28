package telegram

import (
	"context"
	"fmt"
	"time"

	"github.com/krau/ManyACG/internal/infra/config"
	"github.com/krau/ManyACG/internal/intf/telegram/handlers"
	"github.com/krau/ManyACG/internal/intf/telegram/utils"
	"github.com/krau/ManyACG/service"

	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegoapi"
	"github.com/mymmrac/telego/telegohandler"
	"github.com/mymmrac/telego/telegoutil"
)

func InitBot(ctx context.Context) {
	var err error
	apiUrl := config.Get().Telegram.APIURL
	if apiUrl == "" {
		apiUrl = "https://api.telegram.org"
	}
	Bot, err = telego.NewBot(
		config.Get().Telegram.Token,
		telego.WithDefaultLogger(false, true),
		telego.WithAPIServer(apiUrl),
		telego.WithAPICaller(&telegoapi.RetryCaller{
			Caller:       telegoapi.DefaultFastHTTPCaller,
			MaxAttempts:  config.Get().Telegram.Retry.MaxAttempts,
			ExponentBase: config.Get().Telegram.Retry.ExponentBase,
			StartDelay:   time.Duration(config.Get().Telegram.Retry.StartDelay) * time.Second,
			MaxDelay:     time.Duration(config.Get().Telegram.Retry.MaxDelay) * time.Second,
			RateLimit:    telegoapi.RetryRateLimitWaitOrAbort,
		}),
	)
	if err != nil {
		panic(err)
	}

	if config.Get().Telegram.Username != "" {
		ChannelChatID = telegoutil.Username(config.Get().Telegram.Username)
	} else {
		ChannelChatID = telegoutil.ID(config.Get().Telegram.ChatID)
	}
	if ChannelChatID.ID == 0 && ChannelChatID.Username == "" {
		if config.Get().Telegram.Channel {
			panic("telegram channel is enabled, but no channel username or chat ID is set")
		}
	} else {
		IsChannelAvailable = config.Get().Telegram.Channel
	}

	if config.Get().Telegram.GroupID != 0 {
		GroupChatID = telegoutil.ID(config.Get().Telegram.GroupID)
	}
	me, err := Bot.GetMe(ctx)
	if err != nil {
		panic(err)
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
		if config.Get().Telegram.GroupID == 0 {
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
		panic(err)
	}

	botHandler, err := telegohandler.NewBotHandler(Bot, updates)
	if err != nil {
		panic(err)
	}
	go func() {
		<-ctx.Done()
		stopCtx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()
		if err := botHandler.StopWithContext(stopCtx); err != nil {
			fmt.Printf("Error when stopping bot handler: %s\n", err)
		}
	}()

	if !config.Get().Debug {
		botHandler.Use(telegohandler.PanicRecovery())
	}

	baseGroup := botHandler.BaseGroup()
	handlers.RegisterHandlers(baseGroup)
	go func() {
		if err := botHandler.Start(); err != nil {
			fmt.Printf("Error when starting bot handler: %s\n", err)
		}
	}()
	go startService()
}

func startService() {
	go func() {
		for params := range sendArtworkInfoCh {
			if err := utils.SendArtworkInfo(params.Ctx, params.Bot, params.Params); err != nil {
				// common.Logger.Errorf("Error when sending artwork info: %s", err)
				fmt.Printf("Error when sending artwork info: %s\n", err)
			}
			time.Sleep(time.Duration(config.Get().Telegram.Sleep) * time.Second)
		}
	}()
}
