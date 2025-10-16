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
	"github.com/gofiber/fiber/v3/middleware/limiter"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/krau/ManyACG/internal/infra/config/runtimecfg"
	"github.com/krau/ManyACG/internal/interface/rest/common"
	"github.com/krau/ManyACG/internal/interface/rest/handlers"
	"github.com/krau/ManyACG/internal/service"
	"github.com/samber/oops"
)

type RestApp struct {
	fiberApp *fiber.App
	cfg      runtimecfg.RestConfig
	tgbot    common.TelegramBot
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

func New(ctx context.Context, serv *service.Service, cfg runtimecfg.RestConfig, opts ...RestAppOption) (*RestApp, error) {
	errHandler := func(c fiber.Ctx, err error) error {
		c.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSONCharsetUTF8)
		var e *common.Error
		if errors.As(err, &e) {
			if e.Status == fiber.StatusInternalServerError {
				log.Error("internal server error", "err", err, "url", c.OriginalURL())
			}
			return c.Status(e.Status).JSON(e.Response())
		}
		var fe *fiber.Error
		if errors.As(err, &fe) {
			return c.Status(fe.Code).JSON(common.NewError(fe.Code, fe.Message).Response())
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

	app.State().Set(common.StateKeyService, serv)
	app.State().Set(common.StateKeyConfig, cfg)

	if cfg.Limit.Enable {
		app.Use(limiter.New(limiter.Config{
			Expiration: time.Duration(cfg.Limit.Expiration) * time.Second,
			Max:        cfg.Limit.Max,
		}))
	}
	app.Use(cors.New())
	app.Use(compress.New())

	loggerCfg := logger.ConfigDefault
	loggerCfg.Format = "${time} | ${status} | ${latency} | ${ip} | ${method} | ${path} | ${queryParams} | ${error}\n"
	app.Use(logger.New(loggerCfg))

	v1group := app.Group("/api/v1")
	handlers.Register(v1group, serv, cfg)

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
	return restApp, nil
}

func (r *RestApp) Run(stopCtx context.Context) error {
	return r.fiberApp.Listen(r.cfg.Addr, fiber.ListenConfig{
		GracefulContext: stopCtx,
	})
}