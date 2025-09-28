package telegram

import (
	"context"
	"time"

	"github.com/krau/ManyACG/internal/app"
	"github.com/krau/ManyACG/internal/infra/config"
	"github.com/krau/ManyACG/internal/intf/telegram/handlers"
	"github.com/krau/ManyACG/internal/intf/telegram/handlers/shared"
	"github.com/krau/ManyACG/internal/pkg/log"

	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegoapi"
	"github.com/mymmrac/telego/telegohandler"
	"github.com/mymmrac/telego/telegoutil"
)

type botApp struct {
	Bot              *telego.Bot
	Username         string
	ChannelChatID    telego.ChatID
	GroupChatID      telego.ChatID
	ChannelAvailable bool
	app              *app.Application
}

func (b *botApp) RunPolling(ctx context.Context) {
	updates, err := b.Bot.UpdatesViaLongPolling(ctx, &telego.GetUpdatesParams{
		Offset: -1,
		AllowedUpdates: []string{
			telego.MessageUpdates,
			telego.ChannelPostUpdates,
			telego.CallbackQueryUpdates,
			telego.InlineQueryUpdates,
		},
	})
	if err != nil {
		log.Error("failed to get updates via long polling", "err", err)
		return
	}

	botHandler, err := telegohandler.NewBotHandler(b.Bot, updates)
	if err != nil {
		log.Error("error when creating bot handler", "err", err)
		return
	}
	go func() {
		<-ctx.Done()
		stopCtx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()
		if err := botHandler.StopWithContext(stopCtx); err != nil {
			log.Error("error when stopping bot handler", "err", err)
		}
	}()

	if !config.Get().Debug {
		botHandler.Use(telegohandler.PanicRecovery())
	}

	baseGroup := botHandler.BaseGroup()
	handlers.New(&shared.HandlersMeta{
		ChannelChatID:    b.ChannelChatID,
		BotUsername:      b.Username,
		ChannelAvailable: b.ChannelAvailable,
	}, b.app).Register(baseGroup)
	if err := botHandler.Start(); err != nil {
		log.Error("error when starting bot handler", "err", err)
	}
}

func NewBot(ctx context.Context, cfg config.TelegramConfig, app *app.Application) (*botApp, error) {
	apiUrl := cfg.APIURL
	if apiUrl == "" {
		apiUrl = "https://api.telegram.org"
	}
	b, err := telego.NewBot(
		cfg.Token,
		telego.WithDefaultLogger(false, true),
		telego.WithAPIServer(apiUrl),
		telego.WithAPICaller(&telegoapi.RetryCaller{
			Caller:       telegoapi.DefaultFastHTTPCaller,
			MaxAttempts:  cfg.Retry.MaxAttempts,
			ExponentBase: cfg.Retry.ExponentBase,
			StartDelay:   time.Duration(cfg.Retry.StartDelay) * time.Second,
			MaxDelay:     time.Duration(cfg.Retry.MaxDelay) * time.Second,
			RateLimit:    telegoapi.RetryRateLimitWaitOrAbort,
		}),
	)
	if err != nil {
		return nil, err
	}

	var channelChatID, groupChatID telego.ChatID
	var channelAvailable bool
	if cfg.Username != "" {
		channelChatID = telegoutil.Username(cfg.Username)
	} else {
		channelChatID = telegoutil.ID(cfg.ChatID)
	}
	if channelChatID.ID == 0 && channelChatID.Username == "" {
		if cfg.Channel {
			log.Fatal("telegram channel is enabled, but no channel username or chat ID is set")
		}
	} else {
		channelAvailable = cfg.Channel
	}

	if cfg.GroupID != 0 {
		groupChatID = telegoutil.ID(cfg.GroupID)
	}
	me, err := b.GetMe(ctx)
	if err != nil {
		log.Fatal("failed to get bot info, please check your telegram bot token", "err", err)
	}
	botUsername := me.Username

	// utils.Init(channelChatID, groupChatID, botUsername, NewTelegram())

	b.SetMyCommands(ctx, &telego.SetMyCommandsParams{
		Commands: CommonCommands,
		Scope:    &telego.BotCommandScopeDefault{Type: telego.ScopeTypeDefault},
	})

	return &botApp{
		Bot:              b,
		Username:         botUsername,
		ChannelChatID:    channelChatID,
		GroupChatID:      groupChatID,
		ChannelAvailable: channelAvailable,
	}, nil

	// allCommands := append(CommonCommands, AdminCommands...)

	// adminUserIDs, err := service.GetAdminUserIDs(ctx)
	// if err != nil {
	// 	return
	// }

	// for _, adminID := range adminUserIDs {
	// 	Bot.SetMyCommands(ctx, &telego.SetMyCommandsParams{
	// 		Commands: allCommands,
	// 		Scope: &telego.BotCommandScopeChat{
	// 			Type:   telego.ScopeTypeChat,
	// 			ChatID: telegoutil.ID(adminID),
	// 		},
	// 	})
	// 	if config.Get().Telegram.GroupID == 0 {
	// 		continue
	// 	}
	// 	Bot.SetMyCommands(ctx, &telego.SetMyCommandsParams{
	// 		Commands: allCommands,
	// 		Scope: &telego.BotCommandScopeChatMember{
	// 			Type:   telego.ScopeTypeChat,
	// 			ChatID: GroupChatID,
	// 			UserID: adminID,
	// 		},
	// 	})
	// }

	// adminGroupIDs, err := service.GetAdminGroupIDs(ctx)
	// if err != nil {
	// 	return
	// }

	// for _, adminID := range adminGroupIDs {
	// 	Bot.SetMyCommands(ctx, &telego.SetMyCommandsParams{
	// 		Commands: allCommands,
	// 		Scope: &telego.BotCommandScopeChat{
	// 			Type:   telego.ScopeTypeChat,
	// 			ChatID: telegoutil.ID(adminID),
	// 		},
	// 	})
	// }
}

// func startService() {
// 	go func() {
// 		for params := range sendArtworkInfoCh {
// 			if err := utils.SendArtworkInfo(params.Ctx, params.Bot, params.Params); err != nil {
// 				log.Error("failed to send artwork info", "err", err)
// 			}
// 			time.Sleep(time.Duration(config.Get().Telegram.Sleep) * time.Second)
// 		}
// 	}()
// }
