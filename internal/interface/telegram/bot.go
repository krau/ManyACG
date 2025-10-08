package telegram

import (
	"context"
	"time"

	"github.com/krau/ManyACG/internal/infra/config/runtimecfg"
	"github.com/krau/ManyACG/internal/interface/telegram/handlers"
	"github.com/krau/ManyACG/internal/interface/telegram/metautil"
	"github.com/krau/ManyACG/pkg/log"
	"github.com/krau/ManyACG/service"
	"github.com/samber/oops"

	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegoapi"
	"github.com/mymmrac/telego/telegohandler"
	"github.com/mymmrac/telego/telegoutil"
)

type BotApp struct {
	Bot              *telego.Bot
	botUsername      string // 没有 @
	channelChatID    telego.ChatID
	groupChatID      telego.ChatID // 附属群组
	channelAvailable bool          // 是否可以发布到频道
}

func Init(ctx context.Context, serv *service.Service) (*BotApp, error) {
	log.Info("Initialize telegram client")
	cfg := runtimecfg.Get().Telegram
	var err error
	apiUrl := cfg.APIURL
	bot, err := telego.NewBot(
		cfg.Token,
		telego.WithLogger(log.New(log.Config{
			Level:     log.LevelError,
			FileLevel: log.LevelError,
			LogFile:   "logs/telegram.log",
		})),
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
		return nil, oops.Errorf("Error when creating bot: %s", err)
	}
	var channelChatID telego.ChatID
	if cfg.Username != "" {
		channelChatID = telegoutil.Username(cfg.Username)
	} else {
		channelChatID = telegoutil.ID(cfg.ChatID)
	}
	var channelAvailable bool
	if channelChatID.ID == 0 && channelChatID.Username == "" {
		if cfg.Channel {
			return nil, oops.New("Enabled channel mode but no channel ID or username is provided")
		}
	} else {
		channelAvailable = cfg.Channel
	}
	var groupChatID telego.ChatID
	if cfg.GroupID != 0 {
		groupChatID = telegoutil.ID(cfg.GroupID)
	}
	me, err := bot.GetMe(ctx)
	if err != nil {
		log.Fatalf("Error when getting bot info: %s", err)
	}
	botUsername := me.Username

	go func() {
		// set bot commands
		bot.SetMyCommands(ctx, &telego.SetMyCommandsParams{
			Commands: CommonCommands,
			Scope:    &telego.BotCommandScopeDefault{Type: telego.ScopeTypeDefault},
		})
		allCommands := append(CommonCommands, AdminCommands...)
		adminUserIDs, err := serv.GetAdminUserIDs(ctx)
		if err != nil {
			log.Warnf("Error when getting admin user IDs: %s", err)
		} else {
			for _, adminID := range adminUserIDs {
				bot.SetMyCommands(ctx, &telego.SetMyCommandsParams{
					Commands: allCommands,
					Scope: &telego.BotCommandScopeChat{
						Type:   telego.ScopeTypeChat,
						ChatID: telegoutil.ID(adminID),
					},
				})
				if cfg.GroupID == 0 {
					continue
				}
				bot.SetMyCommands(ctx, &telego.SetMyCommandsParams{
					Commands: allCommands,
					Scope: &telego.BotCommandScopeChatMember{
						Type:   telego.ScopeTypeChat,
						ChatID: groupChatID,
						UserID: adminID,
					},
				})
			}
			for _, adminID := range adminUserIDs {
				bot.SetMyCommands(ctx, &telego.SetMyCommandsParams{
					Commands: allCommands,
					Scope: &telego.BotCommandScopeChat{
						Type:   telego.ScopeTypeChat,
						ChatID: telegoutil.ID(adminID),
					},
				})
				if cfg.GroupID == 0 {
					continue
				}
				bot.SetMyCommands(ctx, &telego.SetMyCommandsParams{
					Commands: allCommands,
					Scope: &telego.BotCommandScopeChatMember{
						Type:   telego.ScopeTypeChat,
						ChatID: groupChatID,
						UserID: adminID,
					},
				})
			}
		}
		adminGroupIDs, err := serv.GetAdminGroupIDs(ctx)
		if err != nil {
			log.Warnf("Error when getting admin group IDs: %s", err)
		} else {
			for _, adminID := range adminGroupIDs {
				bot.SetMyCommands(ctx, &telego.SetMyCommandsParams{
					Commands: allCommands,
					Scope: &telego.BotCommandScopeChat{
						Type:   telego.ScopeTypeChat,
						ChatID: telegoutil.ID(adminID),
					},
				})
			}
		}
	}()

	return &BotApp{
		Bot:              bot,
		channelChatID:    channelChatID,
		groupChatID:      groupChatID,
		botUsername:      botUsername,
		channelAvailable: channelAvailable,
	}, nil
}

func (app *BotApp) Run(ctx context.Context, serv *service.Service) {
	log.Info("Start polling")
	updates, err := app.Bot.UpdatesViaLongPolling(ctx, &telego.GetUpdatesParams{
		Offset: -1,
		AllowedUpdates: []string{
			telego.MessageUpdates,
			telego.ChannelPostUpdates,
			telego.CallbackQueryUpdates,
			telego.InlineQueryUpdates,
		},
	})
	if err != nil {
		log.Fatalf("Error when getting updates: %s", err)
	}

	botHandler, err := telegohandler.NewBotHandler(app.Bot, updates)
	if err != nil {
		log.Fatalf("Error when creating bot handler: %s", err)
	}
	go func() {
		<-ctx.Done()
		stopCtx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()
		if err := botHandler.StopWithContext(stopCtx); err != nil {
			log.Warnf("Error when stopping bot handler: %s", err)
		}
		log.Info("Stopped bot handler")
	}()

	if !runtimecfg.Get().App.Debug {
		botHandler.Use(telegohandler.PanicRecoveryHandler(func(recovered any) error {
			log.Errorf("Panic recovered: %v", recovered)
			return nil
		}))
	}
	botHandler.Use(messageLogger)

	baseGroup := botHandler.BaseGroup()
	handlers.New(metautil.NewMetaData(app.channelChatID, app.botUsername), serv).Register(baseGroup)
	if err := botHandler.Start(); err != nil {
		log.Fatalf("Error when starting bot handler: %s", err)
	}
	<-ctx.Done()
	log.Info("Shutting down telegram bot...")
	stopCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second) // [TODO] config
	defer cancel()
	if err := botHandler.StopWithContext(stopCtx); err != nil {
		log.Warnf("Error when stopping bot handler: %s", err)
	}
}
