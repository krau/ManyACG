package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/krau/ManyACG/internal/infra"
	"github.com/krau/ManyACG/internal/infra/config/runtimecfg"
	"github.com/krau/ManyACG/internal/infra/database"
	"github.com/krau/ManyACG/internal/infra/eventbus"
	"github.com/krau/ManyACG/internal/infra/imseek"
	"github.com/krau/ManyACG/internal/infra/internalrpc"
	"github.com/krau/ManyACG/internal/infra/kvstor"
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
	cfg            runtimecfg.Config
	closer         func() error
	restOpts       []rest.RestAppOption
	service        *service.Service
	kv             kvstor.KVStore
	poster         scheduler.ArtworkPoster
	tgbot          restcommon.TelegramBot
	internalServer *internalrpc.Server
	internalAgent  *internalrpc.Agent
}

func NewRuntime(ctx context.Context, cfg runtimecfg.Config) (*Runtime, error) {
	writable := cfg.App.Capabilities.Has(runtimecfg.CapabilityWrite)
	readable := cfg.App.Capabilities.Has(runtimecfg.CapabilityRead)
	if !writable {
		cfg.Telegram.Disable = true
		cfg.Scheduler.Enable = false
		cfg.Imseek.Enable = false
	}
	// 清理上次异常退出(如被强杀)残留在缓存目录的临时文件
	if removed, err := osutil.RemoveStaleTempFiles(cfg.Storage.CacheDir, time.Hour); err != nil {
		log.Warn("failed to remove stale temp files", "dir", cfg.Storage.CacheDir, "err", err)
	} else if removed > 0 {
		log.Info("removed stale temp files", "dir", cfg.Storage.CacheDir, "count", removed)
	}

	resources, err := infra.Init(ctx, cfg)
	if err != nil {
		return nil, err
	}
	closer := resources.Close
	runtime := &Runtime{
		cfg:            cfg,
		closer:         closer,
		kv:             resources.KV,
		internalServer: resources.InternalServer,
		internalAgent:  resources.Agent,
	}
	if !readable && !writable {
		return runtime, nil
	}

	var repos repo.Repositories = database.Default()
	searcher, indexer, err := search.New(ctx, cfg.Search, cfg.App.Capabilities, repos.Artwork())
	if err != nil {
		return nil, errors.Join(fmt.Errorf("initialize search: %w", err), closer())
	}
	// local feature-point search(Optional)
	imCfg := cfg.Imseek
	var eng imseek.Engine
	if imCfg.Enable {
		eng, err = imseek.Init(ctx, imseek.Config{
			Enable:           imCfg.Enable,
			DataDir:          imCfg.DataDir,
			Distance:         imCfg.Distance,
			Count:            imCfg.Count,
			K:                imCfg.K,
			NProbe:           imCfg.NProbe,
			NFeatures:        imCfg.NFeatures,
			MaxHeight:        imCfg.MaxHeight,
			MaxWidth:         imCfg.MaxWidth,
			AutoBuild:        imCfg.AutoBuild,
			BuildDebounceSec: imCfg.BuildDebounceSec,
			MinMatches:       imCfg.MinMatches,
			MinScore:         imCfg.MinScore,
		})
		if err != nil {
			log.Error("imseek init failed, continuing without feature search", "err", err)
			eng, _ = imseek.Init(ctx, imseek.DefaultConfig())
		}
		if eng != nil && eng.Enabled() {
			log.Info("imseek feature search enabled", "data_dir", imCfg.DataDir)
			oldCloser := closer
			closer = func() error {
				return errors.Join(eng.Close(), oldCloser())
			}
		}
	}

	var artworkBus *eventbus.EventBus[*dtoArtworkEventItem]
	needBus := indexer != nil || (eng != nil && eng.Enabled())
	if needBus {
		artworkBus = eventbus.New[*dtoArtworkEventItem]()
		if indexer != nil {
			registerArtworkEventSearcherHandlers(ctx, artworkBus, indexer)
		}
		repos = repo.NewWithArtworkEventImpl(repos, artworkBus)
	}

	var tagger tagging.Tagger
	if writable {
		tagger = tagging.Default()
	}

	servOpts := []service.Option{
		service.WithImseek(eng),
		service.WithIndexer(indexer),
		service.WithArtworkFetcher(resources.Dispatcher),
		service.WithDownloader(resources.Dispatcher),
	}
	if writable {
		servOpts = append(servOpts, service.WithSourcePollers(source.Pollers()))
	}
	serv := service.NewService(
		repos,
		searcher,
		tagger,
		storage.Storages(),
		resources.SourceNamers,
		cfg.Storage,
		servOpts...,
	)
	if artworkBus != nil {
		registerArtworkEventImseekHandlers(ctx, artworkBus, serv)
	}

	var restOpts []rest.RestAppOption
	if readable {
		restOpts = append(restOpts, rest.WithReadService(serv))
	}
	if writable {
		restOpts = append(restOpts, rest.WithWriteService(serv, resources.KV))
		recoverCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		err := serv.RecoverPosting(recoverCtx)
		cancel()
		if err != nil {
			return nil, errors.Join(fmt.Errorf("recover posting artworks: %w", err), closer())
		}
	}
	runtime.closer = closer
	runtime.restOpts = append(runtime.restOpts, restOpts...)
	runtime.service = serv
	return runtime, nil
}

func (r *Runtime) Start(ctx context.Context, stop func()) error {
	if !r.cfg.Rest.Enable && r.cfg.Telegram.Disable && r.internalServer == nil && r.internalAgent == nil {
		return fmt.Errorf("no interfaces enabled: enable REST, Telegram, the internal gateway or an internal agent")
	}
	if r.internalServer != nil {
		go func() {
			if err := r.internalServer.Run(ctx); err != nil {
				log.Error(err)
				stop()
			}
		}()
	}
	if r.internalAgent != nil {
		go r.internalAgent.Run(ctx)
	}
	if !r.cfg.Telegram.Disable {
		botapp, err := telegram.Init(ctx, r.service, r.kv, r.cfg.Telegram, r.cfg.App.Debug)
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
		opts := r.restOpts
		if r.tgbot != nil {
			opts = append(opts, rest.WithTelegramBot(r.tgbot))
		}
		restApp, err := rest.New(r.cfg.Rest, opts...)
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
	return nil
}
