package rest

import (
	"context"
	"errors"
	"time"

	"github.com/charmbracelet/log"
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
}

func New(ctx context.Context, serv *service.Service, cfg runtimecfg.RestConfig) (*RestApp, error) {
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
		ErrorHandler:    errHandler,
		StructValidator: NewStructValidator(),
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
	app.Use(logger.New())

	if serv == nil {
		return nil, oops.New("service is nil")
	}

	v1group := app.Group("/api/v1")
	handlers.Register(v1group, serv, cfg)

	return &RestApp{
		fiberApp: app,
		cfg:      cfg,
	}, nil
}

func (r *RestApp) Run(stopCtx context.Context) error {
	return r.fiberApp.Listen(r.cfg.Addr, fiber.ListenConfig{
		GracefulContext: stopCtx,
	})
}

/*
Api 代码编写备忘
- 请求结构体使用 `message` 标签定义错误时的返回信息
- 校验错误直接在 handler 中原样返回, 外层 errHandler 会处理并返回 BadRequest 和错误信息
- 内部业务也直接返回, 外层 errHandler 会处理并返回 InternalServerError , 并打印日志
- 所有 json 返回值均使用 common.NewSuccess 和 common.NewError 包装
*/