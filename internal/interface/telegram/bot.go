package telegram

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/krau/ManyACG/internal/infra/config/runtimecfg"
	"github.com/krau/ManyACG/internal/infra/kvstor"
	"github.com/krau/ManyACG/internal/interface/telegram/handlers"
	"github.com/krau/ManyACG/internal/interface/telegram/metautil"
	"github.com/krau/ManyACG/internal/service"
	"github.com/krau/ManyACG/internal/shared"
	"github.com/krau/ManyACG/internal/shared/errs"
	"github.com/krau/ManyACG/pkg/log"
	"github.com/samber/oops"

	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegoapi"
	"github.com/mymmrac/telego/telegohandler"
	"github.com/mymmrac/telego/telegoutil"
)

type BotApp struct {
	bot  *telego.Bot
	cfg  runtimecfg.TelegramConfig
	serv *service.Service
	meta *metautil.MetaData
}

func (app *BotApp) Bot() *telego.Bot {
	return app.bot
}

func Init(ctx context.Context, serv *service.Service, cfg runtimecfg.TelegramConfig) (*BotApp, error) {
	log.Info("Initing telegram client")
	var err error
	apiUrl := cfg.APIURL
	bot, err := telego.NewBot(
		cfg.BotToken,
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
	var groupChatID telego.ChatID
	if cfg.GroupID != 0 {
		groupChatID = telegoutil.ID(cfg.GroupID)
	}

	// key: telegram:bot:username:<bot_id>
	// value: bot username without @
	botIdStr := strings.Split(cfg.BotToken, ":")[0]
	botId, err := strconv.Atoi(botIdStr)
	if err != nil {
		return nil, oops.Errorf("Invalid bot token: %s", err)
	}
	key := fmt.Sprintf("telegram:bot:username:%d", botId)

	botUsername, err := kvstor.Get[string](key)
	if err != nil || botUsername == "" {
		me, err := bot.GetMe(ctx)
		if err != nil {
			log.Fatalf("Error when getting bot info: %s", err)
		}
		botUsername = me.Username
	}
	kvstor.Set(key, botUsername)

	admins := cfg.Admins
	for _, adminID := range admins {
		_, err := serv.GetAdminByTelegramID(ctx, adminID)
		if err != nil && !errors.Is(err, errs.ErrRecordNotFound) {
			log.Warnf("Error when getting admin %d: %s", adminID, err)
			continue
		}
		if err == nil {
			continue
		}
		err = serv.CreateAdmin(ctx, adminID, []shared.Permission{shared.PermissionSudo})
		if err != nil {
			log.Warnf("Error when creating admin %d: %s", adminID, err)
			continue
		}
	}

	go func() {
		sig, err := commandsSignature(cfg)
		if err != nil {
			log.Warnf("Error when calculating commands signature: %s", err)
			return
		}
		sigKey := fmt.Sprintf("telegram:bot:commands:%d", botId)
		oldSig, err := kvstor.Get[string](sigKey)
		if err != nil && !errors.Is(err, errs.ErrRecordNotFound) {
			log.Warnf("Error when getting commands signature: %s", err)
			return
		}
		if sig == oldSig {
			// unchanged, skip
			return
		}
		log.Info("Commands signature changed, updating commands...")
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

		err = kvstor.Set(sigKey, sig)
		if err != nil {
			log.Warnf("Error when setting commands signature: %s", err)
			return
		}
	}()

	return &BotApp{
		bot:  bot,
		serv: serv,
		meta: metautil.NewMetaData(channelChatID, botUsername, metautil.WithGroupChatID(groupChatID)),
		cfg:  cfg,
	}, nil
}

func (app *BotApp) Run(ctx context.Context, serv *service.Service) {
	log.Info("Start polling")
	updates, err := app.Bot().UpdatesViaLongPolling(ctx, &telego.GetUpdatesParams{
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

	botHandler, err := telegohandler.NewBotHandler(app.Bot(), updates)
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
	handlers.New(app.meta, serv).Register(baseGroup)
	if err := botHandler.Start(); err != nil {
		log.Fatalf("Error when starting bot handler: %s", err)
	}
	<-ctx.Done()
	log.Info("Shutting down telegram bot...")
	stopCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := botHandler.StopWithContext(stopCtx); err != nil {
		log.Warnf("Error when stopping bot handler: %s", err)
	}
}
