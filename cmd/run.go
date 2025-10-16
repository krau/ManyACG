package cmd

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/duke-git/lancet/v2/retry"
	"github.com/krau/ManyACG/internal/common/version"
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
	"github.com/krau/ManyACG/internal/model/converter"
	"github.com/krau/ManyACG/internal/model/dto"
	"github.com/krau/ManyACG/internal/repo"
	"github.com/krau/ManyACG/internal/service"
	"github.com/krau/ManyACG/pkg/log"
	"github.com/krau/ManyACG/pkg/objectuuid"
	"github.com/krau/ManyACG/pkg/osutil"
)

const banner = `
  __  __                              _       ____    ____ 
 |  \/  |   __ _   _ __    _   _     / \     / ___|  / ___|
 | |\/| |  / _  | | '_ \  | | | |   / _ \   | |     | |  _ 
 | |  | | | (_| | | | | | | |_| |  / ___ \  | |___  | |_| |
 |_|  |_|  \__,_| |_| |_|  \__, | /_/   \_\  \____|  \____|
                           |___/                                        

Build time: %s  Version: %s  Commit: %s
Github: https://github.com/krau/ManyACG
Kawaii is All You Need! ᕕ(◠ڼ◠)ᕗ

`

func Run() {
	fmt.Printf(banner, version.BuildTime, version.Version, version.Commit[:7])
	cfg := runtimecfg.Get()

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

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	log.Info("Starting...")
	closer, err := infra.Init(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if closer != nil {
			if err := closer(); err != nil {
				log.Error(err)
			}
		}
	}()

	osutil.SetCacheTTL(time.Duration(cfg.Storage.CacheTTL) * time.Second)
	osutil.SetOnRemoveError(func(path string, err error) {
		log.Error("remove cache file error", "path", path, "err", err)
	})

	dbRepo := database.Default()

	var repos repo.Repositories
	repos = dbRepo

	if search.Enabled() {
		artworkBus := eventbus.New[*dto.ArtworkEventItem]()
		searcher := search.Default(ctx)
		registerArtworkEventSearcherHandlers(ctx, artworkBus, searcher)
		repos = repo.NewWithArtworkEventImpl(dbRepo, dbRepo, dbRepo, dbRepo, repo.NewArtworkWithEvent(dbRepo, artworkBus), dbRepo, dbRepo, dbRepo, dbRepo, dbRepo, artworkBus)
	}

	serv := service.NewService(
		repos,
		search.Default(ctx),
		tagging.Default(),
		storage.Storages(),
		source.Sources(),
		cfg.Storage)
	service.SetDefault(serv)

	var poster scheduler.ArtworkPoster
	var restTgBotService restcommon.TelegramBot
	if !cfg.Telegram.Disable {
		botapp, err := telegram.Init(ctx, serv, cfg.Telegram)
		if err != nil {
			log.Fatal(err)
		}
		go botapp.Run(ctx, serv)
		poster = botapp
		restTgBotService = botapp
	}
	if cfg.Scheduler.Enable && poster != nil {
		go scheduler.StartPoster(ctx, poster, serv)
	}
	if cfg.Rest.Enable {
		opts := []rest.RestAppOption{}
		if restTgBotService != nil {
			opts = append(opts, rest.WithTelegramBot(restTgBotService))
		}
		restApp, err := rest.New(ctx, serv, cfg.Rest, opts...)
		if err != nil {
			log.Fatal(err)
		}
		go func() {
			log.Info("Starting RESTful API server", "addr", cfg.Rest.Addr)
			if err := restApp.Run(ctx); err != nil {
				log.Error(err)
				stop()
			}
		}()
	}

	log.Info("ManyACG is running !")

	defer log.Info("Exited.")
	<-ctx.Done()
	cleanCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := serv.Cleanup(cleanCtx); err != nil {
		log.Error(err)
	}
}

func registerArtworkEventSearcherHandlers(ctx context.Context, bus repo.EventBus[*dto.ArtworkEventItem], searcher search.Searcher) {
	filter := func(payload *dto.ArtworkEventItem) bool {
		return payload != nil && payload.ID != objectuuid.Nil
	}
	bus.Subscribe(repo.EventTypeArtworkCreate, func(payload *dto.ArtworkEventItem) {
		retry.Retry(func() error {
			doc := converter.DtoArtworkEventItemToSearchDocument(payload)
			if doc == nil {
				return nil
			}
			err := searcher.AddDocuments(ctx, []*dto.ArtworkSearchDocument{doc})
			if err != nil {
				log.Error(err)
				return err
			}
			log.Debug("indexed artwork", "id", payload.ID, "title", payload.Title)
			return nil
		}, retry.Context(ctx))
	}, filter)
	bus.Subscribe(repo.EventTypeArtworkUpdate, func(payload *dto.ArtworkEventItem) {
		retry.Retry(func() error {
			doc := converter.DtoArtworkEventItemToSearchDocument(payload)
			if doc == nil {
				return nil
			}
			err := searcher.AddDocuments(ctx, []*dto.ArtworkSearchDocument{doc})
			if err != nil {
				log.Error(err)
				return err
			}
			log.Debug("re-indexed artwork", "id", payload.ID, "title", payload.Title)
			return nil
		}, retry.Context(ctx))
	}, filter)
	bus.Subscribe(repo.EventTypeArtworkDelete, func(payload *dto.ArtworkEventItem) {
		retry.Retry(func() error {
			err := searcher.DeleteDocuments(ctx, []string{payload.ID.Hex()})
			if err != nil {
				log.Error(err)
				return err
			}
			log.Debug("deleted indexed artwork", "id", payload.ID, "title", payload.Title)
			return nil
		}, retry.Context(ctx))
	}, filter)
}
