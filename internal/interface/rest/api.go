package rest

import (
	"context"
	"errors"
	"time"

	"github.com/charmbracelet/log"
	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/compress"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/etag"
	"github.com/gofiber/fiber/v3/middleware/limiter"
	"github.com/gofiber/fiber/v3/middleware/logger"
	recoverer "github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/krau/ManyACG/internal/infra/config/runtimecfg"
	"github.com/krau/ManyACG/internal/infra/kvstor"
	"github.com/krau/ManyACG/internal/interface/rest/common"
	"github.com/krau/ManyACG/internal/interface/rest/handlers"
	"github.com/krau/ManyACG/internal/shared/errs"

	"github.com/samber/oops"
)

type RestApp struct {
	fiberApp        *fiber.App
	cfg             runtimecfg.RestConfig
	tgbot           common.TelegramBot
	writer          common.WriteService
	store           kvstor.KVStore
	fetcher         common.ArtworkFetcher
	downloader      common.ArtworkDownloader
	internalToken   string
	internalTimeout time.Duration
}

type queryService struct {
	common.QueryService
}

type RestAppOption func(app *RestApp) error

func WithTelegramBot(bot common.TelegramBot) RestAppOption {
	return func(app *RestApp) error {
		if bot == nil {
			return oops.New("telegram bot is nil")
		}
		app.tgbot = bot
		return nil
	}
}

func WithWriteService(serv common.WriteService, store kvstor.KVStore) RestAppOption {
	return func(app *RestApp) error {
		if serv == nil {
			return oops.New("write service is nil")
		}
		if store == nil {
			return oops.New("write KV store is nil")
		}
		app.writer = serv
		app.store = store
		return nil
	}
}

func WithArtworkFetcher(fetcher common.ArtworkFetcher, token string, timeout time.Duration) RestAppOption {
	return func(app *RestApp) error {
		if fetcher == nil {
			return oops.New("artwork fetcher is nil")
		}
		if err := app.setInternalAuth(token, timeout); err != nil {
			return err
		}
		app.fetcher = fetcher
		return nil
	}
}

func WithArtworkDownloader(downloader common.ArtworkDownloader, token string, timeout time.Duration) RestAppOption {
	return func(app *RestApp) error {
		if downloader == nil {
			return oops.New("artwork downloader is nil")
		}
		if err := app.setInternalAuth(token, timeout); err != nil {
			return err
		}
		app.downloader = downloader
		return nil
	}
}

func (app *RestApp) setInternalAuth(token string, timeout time.Duration) error {
	if token == "" {
		return oops.New("internal token is required")
	}
	if timeout <= 0 {
		return oops.New("internal request timeout must be positive")
	}
	app.internalToken = token
	app.internalTimeout = timeout
	return nil
}

func New(ctx context.Context, serv common.QueryService, cfg runtimecfg.RestConfig, opts ...RestAppOption) (*RestApp, error) {
	errHandler := func(c fiber.Ctx, err error) error {
		c.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSONCharsetUTF8)
		if e, ok := errors.AsType[*common.Error](err); ok {
			if e.Status == fiber.StatusInternalServerError {
				log.Error("internal server error", "err", err, "url", c.OriginalURL())
			}
			return c.Status(e.Status).JSON(e.Response())
		}
		if fe, ok := errors.AsType[*fiber.Error](err); ok {
			return c.Status(fe.Code).JSON(common.NewError(fe.Code, fe.Message).Response())
		}
		if errors.Is(err, errs.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(common.NewError(fiber.StatusNotFound, "resource not found").Response())
		}
		log.Error("internal server error", "err", err, "url", c.OriginalURL())
		code := fiber.StatusInternalServerError
		return c.Status(code).JSON(common.NewError(code, "internal server error").Response())
	}

	app := fiber.New(fiber.Config{
		JSONEncoder:     json.Marshal,
		JSONDecoder:     json.Unmarshal,
		ErrorHandler:    errHandler,
		StructValidator: NewStructValidator(),
		TrustProxy:      true,
		TrustProxyConfig: fiber.TrustProxyConfig{
			LinkLocal: true,
			Private:   true,
			Loopback:  true,
		},
		ProxyHeader:        fiber.HeaderXForwardedFor,
		EnableIPValidation: true,
	})

	app.State().Set(common.StateKeyQueryService, queryService{QueryService: serv})
	app.State().Set(common.StateKeyConfig, cfg)

	if cfg.Limit.Enable {
		app.Use(limiter.New(limiter.Config{
			Expiration: time.Duration(cfg.Limit.Expiration) * time.Second,
			Max:        cfg.Limit.Max,
		}))
	}
	app.Use(cors.New())
	app.Use(etag.New())
	app.Use(compress.New())
	app.Use(recoverer.New())

	loggerCfg := logger.ConfigDefault
	loggerCfg.Format = "${time} | ${status} | ${latency} | ${ip} | ${method} | ${path} | ${queryParams} | ${error}\n"
	app.Use(logger.New(loggerCfg))

	restApp := &RestApp{
		fiberApp: app,
		cfg:      cfg,
	}

	for _, opt := range opts {
		if err := opt(restApp); err != nil {
			return nil, oops.Wrapf(err, "applying option")
		}
	}
	if restApp.tgbot != nil {
		app.State().Set(common.StateKeyTelegramBot, restApp.tgbot)
	}
	if restApp.writer != nil {
		app.State().Set(common.StateKeyWriteService, restApp.writer)
		app.State().Set(common.StateKeyKVStore, restApp.store)
	}
	root := app.Group("/api/v1")
	handlers.Register(root, cfg, restApp.writer)
	if restApp.fetcher != nil {
		root.Post("/internal/artwork/fetch", handlers.NewInternalArtworkFetchHandler(restApp.fetcher, restApp.internalToken, restApp.internalTimeout))
	}
	if restApp.downloader != nil {
		root.Post("/internal/artwork/download", handlers.NewInternalArtworkDownloadHandler(restApp.downloader, restApp.internalToken, restApp.internalTimeout))
	}
	return restApp, nil
}

func (r *RestApp) Run(stopCtx context.Context) error {
	return r.fiberApp.Listen(r.cfg.Addr, fiber.ListenConfig{
		GracefulContext: stopCtx,
	})
}
