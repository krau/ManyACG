package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
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
	service        *service.Service
	kv             kvstor.KVStore
	internalServer *internalrpc.Server
	internalAgent  *internalrpc.Agent
	stop           func()
	interfaces     sync.WaitGroup
}

func NewRuntime(ctx context.Context, cfg runtimecfg.Config) (*Runtime, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	mainNode := cfg.App.IsMain()
	botEnabled := mainNode && !cfg.Telegram.Disable
	if !cfg.Rest.Enable && !botEnabled && cfg.Internal.Gateway.Addr == "" && cfg.Internal.Agent.Gateway == "" {
		return nil, fmt.Errorf("no interfaces enabled: enable REST, Telegram, the internal gateway or an internal agent")
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
	if !mainNode && !cfg.App.Capabilities.Has(runtimecfg.CapabilityRead) {
		return runtime, nil
	}

	var repos repo.Repositories = database.Default()
	searcher, indexer, err := search.New(ctx, cfg.Search, repos.Artwork(), mainNode)
	if err != nil {
		return nil, errors.Join(fmt.Errorf("initialize search: %w", err), closer())
	}
	// local feature-point search(Optional)
	imCfg := cfg.Imseek
	var eng imseek.Engine
	if mainNode && imCfg.Enable {
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
	if mainNode {
		tagger = tagging.Default()
	}

	servOpts := []service.Option{
		service.WithImseek(eng),
		service.WithIndexer(indexer),
		service.WithArtworkFetcher(resources.Dispatcher),
		service.WithDownloader(resources.Dispatcher),
	}
	if mainNode {
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

	if mainNode {
		recoverCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		err := serv.RecoverPosting(recoverCtx)
		cancel()
		if err != nil {
			return nil, errors.Join(fmt.Errorf("recover posting artworks: %w", err), closer())
		}
	}
	runtime.closer = closer
	runtime.service = serv
	return runtime, nil
}

func (r *Runtime) goInterface(fn func()) {
	r.interfaces.Add(1)
	go func() {
		defer r.interfaces.Done()
		fn()
	}()
}

func (r *Runtime) Start(ctx context.Context, stop func()) error {
	r.stop = stop
	botEnabled := r.cfg.App.IsMain() && !r.cfg.Telegram.Disable
	if r.internalServer != nil {
		r.goInterface(func() {
			if err := r.internalServer.Run(ctx); err != nil {
				log.Error(err)
				stop()
			}
		})
	}
	if r.internalAgent != nil {
		r.goInterface(func() { r.internalAgent.Run(ctx) })
	}
	var bot restcommon.TelegramBot
	if botEnabled {
		botapp, err := telegram.Init(ctx, r.service, r.kv, r.cfg.Telegram, r.cfg.App.Debug)
		if err != nil {
			stop()
			return err
		}
		r.goInterface(func() {
			botapp.Run(ctx, r.service)
			botapp.Wait()
		})
		bot = botapp
		if r.cfg.Scheduler.Enable {
			r.goInterface(func() { scheduler.StartPoster(ctx, r.cfg.Scheduler, botapp, r.service) })
		}
	}

	if r.cfg.Rest.Enable {
		var opts []rest.RestAppOption
		if bot != nil {
			opts = append(opts, rest.WithTelegramBot(bot))
		}
		restApp, err := rest.New(r.cfg.Rest, r.cfg.App, r.service, r.kv, opts...)
		if err != nil {
			stop()
			return err
		}
		r.goInterface(func() {
			log.Info("Starting RESTful API server", "addr", r.cfg.Rest.Addr)
			if err := restApp.Run(ctx); err != nil {
				log.Error(err)
				stop()
			}
		})
	}

	return nil
}

func (r *Runtime) Close() error {
	if r.stop != nil {
		r.stop()
	}
	r.waitInterfaces()
	if r.closer == nil {
		return nil
	}
	return r.closer()
}

const interfaceReturnSlack = time.Second

func (r *Runtime) waitInterfaces() {
	budget := r.cfg.App.ShutdownDuration() + interfaceReturnSlack
	done := make(chan struct{})
	go func() {
		r.interfaces.Wait()
		close(done)
	}()
	timer := time.NewTimer(budget)
	defer timer.Stop()
	select {
	case <-done:
	case <-timer.C:
		log.Warn("interfaces did not stop in time, closing resources anyway", "budget", budget)
	}
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
