package cmd

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/krau/ManyACG/internal/common/version"
	_ "github.com/krau/ManyACG/internal/infra"
	"github.com/krau/ManyACG/internal/infra/config/runtimecfg"
	"github.com/krau/ManyACG/internal/infra/database"
	"github.com/krau/ManyACG/internal/infra/eventbus"
	"github.com/krau/ManyACG/internal/infra/search"
	"github.com/krau/ManyACG/internal/infra/source"
	"github.com/krau/ManyACG/internal/infra/storage"
	"github.com/krau/ManyACG/internal/infra/tagging"
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

	osutil.SetCacheTTL(time.Duration(cfg.Storage.CacheTTL) * time.Second)
	osutil.SetOnRemoveError(func(path string, err error) {
		log.Error("remove cache file error", "path", path, "err", err)
	})

	source.InitAll()
	if err := storage.InitAll(ctx); err != nil {
		log.Fatal(err)
	}

	database.Init(ctx)
	dbRepo := database.Default()

	var repos repo.Repositories
	repos = dbRepo

	if search.Enabled() {
		artworkBus := eventbus.New[*dto.ArtworkEventItem]()
		searcher := search.Default(ctx)
		registerArtworkEventSearcherHandlers(ctx, artworkBus, searcher)
		repos = &repo.WithArtworkEventImpl{
			Tx:          dbRepo,
			AdminRepo:   dbRepo,
			ApiKeyRepo:  dbRepo,
			ArtistRepo:  dbRepo,
			TagRepo:     dbRepo,
			PictureRepo: dbRepo,
			DeletedRepo: dbRepo,
			CachedRepo:  dbRepo,
			ArtworkRepo: repo.NewArtworkWithEvent(dbRepo, artworkBus),
			ArtworkBus:  artworkBus,
		}
	}

	serv := service.NewService(
		repos,
		search.Default(ctx),
		tagging.Default(),
		storage.Storages(),
		source.Sources(),
		runtimecfg.Get().Storage)
	service.SetDefault(serv)

	botapp, err := telegram.Init(ctx, serv)
	if err != nil {
		log.Fatal(err)
	}
	go botapp.Run(ctx, serv)
	if runtimecfg.Get().Scheduler.Enable {
		go scheduler.StartPoster(ctx, botapp, serv)
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
		doc := converter.DtoArtworkEventItemToSearchDocument(payload)
		if doc == nil {
			return
		}
		err := searcher.AddDocuments(ctx, []*dto.ArtworkSearchDocument{doc})
		if err != nil {
			log.Error(err)
			return
		}
		log.Debug("indexed artwork", "id", payload.ID, "title", payload.Title)
	}, filter)
	bus.Subscribe(repo.EventTypeArtworkUpdate, func(payload *dto.ArtworkEventItem) {
		doc := converter.DtoArtworkEventItemToSearchDocument(payload)
		if doc == nil {
			return
		}
		err := searcher.AddDocuments(ctx, []*dto.ArtworkSearchDocument{doc})
		if err != nil {
			log.Error(err)
			return
		}
		log.Debug("re-indexed artwork", "id", payload.ID, "title", payload.Title)
	}, filter)
	bus.Subscribe(repo.EventTypeArtworkDelete, func(payload *dto.ArtworkEventItem) {
		err := searcher.DeleteDocuments(ctx, []string{payload.ID.Hex()})
		if err != nil {
			log.Error(err)
			return
		}
		log.Debug("deleted indexed artwork", "id", payload.ID, "title", payload.Title)
	}, filter)
}

// func cleanCacheDir(cfg runtimecfg.Config) {
// 	if cfg.Storage.CacheDir != "" && !cfg.App.Debug {
// 		for _, path := range []string{"/", ".", "\\", ".."} {
// 			if filepath.Clean(cfg.Storage.CacheDir) == path {
// 				log.Error("Invalid cache dir: ", cfg.Storage.CacheDir)
// 				return
// 			}
// 		}
// 		currentDir, err := os.Getwd()
// 		if err != nil {
// 			log.Error(err)
// 			return
// 		}
// 		cachePath := filepath.Join(currentDir, cfg.Storage.CacheDir)
// 		cachePath, err = filepath.Abs(cachePath)
// 		if err != nil {
// 			log.Error(err)
// 			return
// 		}
// 		log.Info("Removing cache dir: ", cachePath)
// 		if err := os.RemoveAll(cachePath); err != nil {
// 			log.Error(err)
// 			return
// 		}
// 	}
// }
