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
	"github.com/krau/ManyACG/internal/interface/telegram/handlers/utils"
	"github.com/krau/ManyACG/internal/interface/telegram/metautil"
	"github.com/krau/ManyACG/internal/service"
	"github.com/krau/ManyACG/internal/shared"
	"github.com/krau/ManyACG/internal/shared/errs"
	"github.com/krau/ManyACG/pkg/log"
	telegoapiwrapper "github.com/krau/ManyACG/pkg/telegoapi"
	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegoapi"
	"github.com/mymmrac/telego/telegohandler"
	"github.com/mymmrac/telego/telegoutil"
	"github.com/samber/oops"
)

type BotApp struct {
	bot              *telego.Bot
	serv             *service.Service
	meta             *metautil.MetaData
	cfg              runtimecfg.TelegramConfig
	debug            bool
	artworkInfoQueue chan artworkInfoTask
}

func (app *BotApp) Bot() *telego.Bot {
	return app.bot
}

func Init(ctx context.Context, serv *service.Service, cfg runtimecfg.TelegramConfig, debug bool) (*BotApp, error) {
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
		telego.WithRequestConstructor(telegoapiwrapper.MultipartRequestConstructor{}),
		telego.WithAPICaller(&telegoapiwrapper.RetryRateLimitCaller{
			Caller:       telegoapi.DefaultFastHTTPCaller,
			MaxAttempts:  cfg.Retry.MaxAttempts,
			ExponentBase: cfg.Retry.ExponentBase,
			StartDelay:   time.Duration(cfg.Retry.StartDelay) * time.Second,
			MaxDelay:     time.Duration(cfg.Retry.MaxDelay) * time.Second,
			RateLimit:    telegoapi.RetryRateLimitWait,
		}),
	)
	if err != nil {
		return nil, oops.Errorf("Error when creating bot: %s", err)
	}
	var channelChatID telego.ChatID
	if cfg.ChatID != 0 && cfg.Username != "" {
		channelChatID = telegoutil.ID(cfg.ChatID)
		channelChatID.Username = cfg.Username
	} else if cfg.ChatID != 0 {
		channelChatID = telegoutil.ID(cfg.ChatID)
	} else if cfg.Username != "" {
		channelChatID = telegoutil.Username(cfg.Username)
	} else {
		return nil, oops.New("Either ChatID or Username must be set in config")
	}
	if channelChatID.ID == 0 || channelChatID.Username == "" {
		chatFull, err := bot.GetChat(ctx, &telego.GetChatParams{ChatID: channelChatID})
		if err != nil {
			return nil, oops.Errorf("Error when getting chat info: %s", err)
		}
		channelChatID.ID = chatFull.ID
		channelChatID.Username = chatFull.Username
	}

	var groupChatID telego.ChatID
	if cfg.GroupID != 0 {
		groupChatID = telegoutil.ID(cfg.GroupID)
	}

	// key: telegram:bot:username:<bot_id>
	// value: bot username without @
	botIdStr, _, _ := strings.Cut(cfg.BotToken, ":")
	botId, err := strconv.Atoi(botIdStr)
	if err != nil {
		return nil, oops.Errorf("Invalid bot token: %s", err)
	}
	key := fmt.Sprintf("telegram:bot:username:%d", botId)

	botUsername, err := kvstor.Get[string](ctx, key)
	if err != nil || botUsername == "" {
		me, err := bot.GetMe(ctx)
		if err != nil {
			log.Fatalf("Error when getting bot info: %s", err)
		}
		botUsername = me.Username
	}
	kvstor.Set(ctx, key, botUsername)

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
		oldSig, err := kvstor.Get[string](ctx, sigKey)
		if err != nil && !errors.Is(err, errs.ErrRecordNotFound) {
			log.Warnf("Error when getting commands signature: %s", err)
			return
		}
		if sig == oldSig {
			// unchanged, skip
			return
		}
		log.Info("Commands signature changed, updating commands...")
		setCommands(ctx, bot, CommonCommands, &telego.BotCommandScopeDefault{Type: telego.ScopeTypeDefault})

		allCommands := append(CommonCommands, AdminCommands...)
		adminUserIDs, err := serv.GetAdminUserIDs(ctx)
		if err != nil {
			log.Warnf("Error when getting admin user IDs: %s", err)
		} else {
			syncAdminUserCommands(ctx, bot, allCommands, adminUserIDs, groupChatID)
		}
		adminGroupIDs, err := serv.GetAdminGroupIDs(ctx)
		if err != nil {
			log.Warnf("Error when getting admin group IDs: %s", err)
		} else {
			for _, adminID := range adminGroupIDs {
				setCommands(ctx, bot, allCommands, &telego.BotCommandScopeChat{
					Type:   telego.ScopeTypeChat,
					ChatID: telegoutil.ID(adminID),
				})
			}
		}

		err = kvstor.Set(ctx, sigKey, sig)
		if err != nil {
			log.Warnf("Error when setting commands signature: %s", err)
			return
		}
	}()

	metaopts := []metautil.Option{}
	if cfg.GroupID != 0 {
		metaopts = append(metaopts, metautil.WithGroupChatID(groupChatID))
	}
	metaopts = append(metaopts, metautil.WithBotID(int64(botId)))
	meta := metautil.NewMetaData(channelChatID, botUsername, metaopts...)

	artworkInfoQueue := make(chan artworkInfoTask, 100)
	app := &BotApp{
		bot:              bot,
		serv:             serv,
		meta:             meta,
		cfg:              cfg,
		debug:            debug,
		artworkInfoQueue: artworkInfoQueue,
	}

	go app.processArtworkInfoTasks(ctx)

	return app, nil
}

func setCommands(ctx context.Context, bot *telego.Bot, commands []telego.BotCommand, scope telego.BotCommandScope) {
	if err := bot.SetMyCommands(ctx, &telego.SetMyCommandsParams{Commands: commands, Scope: scope}); err != nil {
		log.Warnf("Error when setting commands for %T: %s", scope, err)
	}
}

func syncAdminUserCommands(ctx context.Context, bot *telego.Bot, commands []telego.BotCommand, adminUserIDs []int64, groupChatID telego.ChatID) {
	for _, adminID := range adminUserIDs {
		setCommands(ctx, bot, commands, &telego.BotCommandScopeChat{
			Type:   telego.ScopeTypeChat,
			ChatID: telegoutil.ID(adminID),
		})
		if groupChatID.ID == 0 {
			continue
		}
		setCommands(ctx, bot, commands, &telego.BotCommandScopeChatMember{
			Type:   telego.ScopeTypeChat,
			ChatID: groupChatID,
			UserID: adminID,
		})
	}
}

func (app *BotApp) processArtworkInfoTasks(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			log.Info("Stopping artwork info task processor")
			return
		case task := <-app.artworkInfoQueue:
			err := utils.SendArtworkInfo(task.ctx, app.bot, app.meta, app.serv, task.sourceUrl, telegoutil.ID(task.chatID), utils.SendArtworkInfoOptions{AppendCaption: task.appendCaption, HasPermission: true})
			if err != nil {
				log.Errorf("Error when sending artwork info: %s", err)
			}
		}
	}
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

	botHandler, err := telegohandler.NewBotHandler(app.Bot(), updates,
		telegohandler.WithErrorHandler(func(ctx *telegohandler.Context, update telego.Update, err error) {
			fields := append([]any{"err", err}, updateLogFields(update)...)
			log.Error("telegram handler error", fields...)
		}),
	)
	if err != nil {
		log.Fatalf("Error when creating bot handler: %s", err)
	}
	go func() {
		<-ctx.Done()
		log.Info("Shutting down telegram bot...")
		stopCtx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()
		if err := botHandler.StopWithContext(stopCtx); err != nil {
			log.Warnf("Error when stopping bot handler: %s", err)
		}
		log.Info("Stopped bot handler")
	}()

	if !app.debug {
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
}
