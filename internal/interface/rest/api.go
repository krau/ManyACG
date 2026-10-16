package rest

import (
	"context"
	"errors"
	"net"
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
	"github.com/krau/ManyACG/internal/service"
	"github.com/krau/ManyACG/internal/shared/errs"

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

func New(cfg runtimecfg.RestConfig, app runtimecfg.AppConfig, serv *service.Service, store kvstor.KVStore, opts ...RestAppOption) (*RestApp, error) {
	if (app.IsMain() || app.Capabilities.Has(runtimecfg.CapabilityRead)) && serv == nil {
		return nil, oops.New("service is required for main or read REST routes")
	}
	if app.IsMain() && store == nil {
		return nil, oops.New("KV store is required for main REST routes")
	}
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

	fiberApp := fiber.New(fiber.Config{
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

	fiberApp.State().Set(common.StateKeyConfig, cfg)
	if serv != nil {
		fiberApp.State().Set(common.StateKeyService, serv)
	}
	if store != nil {
		fiberApp.State().Set(common.StateKeyKVStore, store)
	}

	if cfg.Limit.Enable {
		fiberApp.Use(limiter.New(limiter.Config{
			Expiration: time.Duration(cfg.Limit.Expiration) * time.Second,
			Max:        cfg.Limit.Max,
		}))
	}
	fiberApp.Use(cors.New())
	fiberApp.Use(etag.New())
	fiberApp.Use(compress.New())
	fiberApp.Use(recoverer.New())

	loggerCfg := logger.ConfigDefault
	loggerCfg.Format = "${time} | ${status} | ${latency} | ${ip} | ${method} | ${path} | ${queryParams} | ${error}\n"
	fiberApp.Use(logger.New(loggerCfg))

	restApp := &RestApp{
		fiberApp: fiberApp,
		cfg:      cfg,
	}

	for _, opt := range opts {
		if err := opt(restApp); err != nil {
			return nil, oops.Wrapf(err, "applying option")
		}
	}
	root := fiberApp.Group("/api/v1")
	if restApp.tgbot != nil {
		fiberApp.State().Set(common.StateKeyTelegramBot, restApp.tgbot)
	}
	handlers.Register(root, cfg, app)
	return restApp, nil
}

func (r *RestApp) Run(stopCtx context.Context) error {
	listener, err := net.Listen("tcp4", r.cfg.Addr)
	if err != nil {
		return err
	}
	defer listener.Close()
	// Closing the listener also handles cancellation before Serve registers it.
	stopClosing := context.AfterFunc(stopCtx, func() { _ = listener.Close() })
	defer stopClosing()
	err = r.fiberApp.Listener(listener)
	shutdownErr := r.fiberApp.Shutdown()
	if errors.Is(shutdownErr, net.ErrClosed) {
		shutdownErr = nil
	}
	return errors.Join(err, shutdownErr)
}
