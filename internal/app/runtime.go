package app

import (
	"context"
	"net/http"
	"time"

	"github.com/krau/ManyACG/internal/infra"
	"github.com/krau/ManyACG/internal/infra/config/runtimecfg"
	"github.com/krau/ManyACG/internal/infra/database"
	"github.com/krau/ManyACG/internal/infra/eventbus"
	"github.com/krau/ManyACG/internal/infra/search"
	"github.com/krau/ManyACG/internal/infra/source"
	"github.com/krau/ManyACG/internal/infra/storage"
	"github.com/krau/ManyACG/internal/infra/tagging"
	"github.com/krau/ManyACG/internal/interface/rest"
	restcommon "github.com/krau/ManyACG/internal/interface/rest/common"
	"github.com/krau/ManyACG/internal/interface/scheduler"
	"github.com/krau/ManyACG/internal/interface/telegram"
	"github.com/krau/ManyACG/internal/repo"
	"github.com/krau/ManyACG/internal/service"
	"github.com/krau/ManyACG/pkg/log"
	"github.com/krau/ManyACG/pkg/osutil"
)

type Runtime struct {
	cfg     runtimecfg.Config
	closer  func() error
	repos   repo.Repositories
	search  search.Searcher
	service *service.Service
	poster  scheduler.ArtworkPoster
	tgbot   restcommon.TelegramBot
}

func NewRuntime(ctx context.Context, cfg runtimecfg.Config) (*Runtime, error) {
	closer, err := infra.Init(ctx, cfg)
	if err != nil {
		return nil, err
	}

	dbRepo := database.Default()
	searcher := search.Default(ctx)
	repos := repo.Repositories(dbRepo)
	if search.Enabled() {
		artworkBus := eventbus.New[*dtoArtworkEventItem]()
		registerArtworkEventSearcherHandlers(ctx, artworkBus, searcher)
		repos = repo.NewWithArtworkEventImpl(dbRepo, artworkBus)
	}

	serv := service.NewService(
		repos,
		searcher,
		tagging.Default(),
		storage.Storages(),
		source.Sources(),
		cfg.Storage,
	)

	return &Runtime{
		cfg:     cfg,
		closer:  closer,
		repos:   repos,
		search:  searcher,
		service: serv,
	}, nil
}

func (r *Runtime) Start(ctx context.Context, stop func()) error {
	if !r.cfg.Telegram.Disable {
		botapp, err := telegram.Init(ctx, r.service, r.cfg.Telegram, r.cfg.App.Debug)
		if err != nil {
			return err
		}
		go botapp.Run(ctx, r.service)
		r.poster = botapp
		r.tgbot = botapp
	}

	if r.cfg.Scheduler.Enable && r.poster != nil {
		go scheduler.StartPosterWithConfig(ctx, r.cfg.Scheduler, r.poster, r.service)
	}

	if r.cfg.Rest.Enable {
		opts := []rest.RestAppOption{}
		if r.tgbot != nil {
			opts = append(opts, rest.WithTelegramBot(r.tgbot))
		}
		restApp, err := rest.New(ctx, r.service, r.cfg.Rest, opts...)
		if err != nil {
			return err
		}
		go func() {
			log.Info("Starting RESTful API server", "addr", r.cfg.Rest.Addr)
			if err := restApp.Run(ctx); err != nil {
				log.Error(err)
				stop()
			}
		}()
	}

	return nil
}

func (r *Runtime) Close() error {
	if r.closer == nil {
		return nil
	}
	return r.closer()
}

func (r *Runtime) Cleanup(ctx context.Context) error {
	return r.service.Cleanup(ctx)
}

func Run(ctx context.Context, cfg runtimecfg.Config, stop func()) error {
	log.SetDefault(log.New(log.Config{
		LogFile:    cfg.Log.FilePath,
		MaxBackups: int(cfg.Log.BackupNum),
	}))
	if cfg.App.Debug {
		go func() {
			log.Info("Start pprof server")
			if err := http.ListenAndServe("localhost:39060", nil); err != nil {
				log.Fatal(err)
			}
		}()
	}

	runtime, err := NewRuntime(ctx, cfg)
	if err != nil {
		return err
	}
	defer func() {
		if err := runtime.Close(); err != nil {
			log.Error(err)
		}
	}()

	osutil.SetCacheTTL(time.Duration(cfg.Storage.CacheTTL) * time.Second)
	osutil.SetOnRemoveError(func(path string, err error) {
		log.Error("remove cache file error", "path", path, "err", err)
	})

	if err := runtime.Start(ctx, stop); err != nil {
		return err
	}

	log.Info("ManyACG is running !")
	defer log.Info("Exited.")

	<-ctx.Done()
	cleanCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return runtime.Cleanup(cleanCtx)
}
