package cmd

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/krau/ManyACG/internal/common/version"
	_ "github.com/krau/ManyACG/internal/infra"
	"github.com/krau/ManyACG/internal/infra/config/runtimecfg"
	"github.com/krau/ManyACG/internal/infra/database"
	"github.com/krau/ManyACG/internal/infra/search"
	"github.com/krau/ManyACG/internal/infra/source"
	"github.com/krau/ManyACG/internal/infra/storage"
	"github.com/krau/ManyACG/internal/interface/telegram"
	"github.com/krau/ManyACG/pkg/log"
	"github.com/krau/ManyACG/service"
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

	if runtimecfg.Get().App.Debug {
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
	database.Init(ctx)
	source.InitAll()
	if err := storage.InitAll(ctx); err != nil {
		log.Fatal(err)
	}
	database.Init(ctx)
	serv := service.NewService(database.Default(), search.Default(), storage.Storages(), source.Sources(), runtimecfg.Get().Storage)
	botapp, err := telegram.Init(ctx, serv)
	if err != nil {
		log.Fatal(err)
	}
	go botapp.Run(ctx, serv)

	log.Info("ManyACG is running !")

	defer log.Info("Exited.")
	<-ctx.Done()
	cleanCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := serv.Cleanup(cleanCtx); err != nil {
		log.Error(err)
	}
	// cleanCacheDir(runtimecfg.Get())
}

func cleanCacheDir(cfg runtimecfg.Config) {
	if cfg.Storage.CacheDir != "" && !cfg.App.Debug {
		for _, path := range []string{"/", ".", "\\", ".."} {
			if filepath.Clean(cfg.Storage.CacheDir) == path {
				log.Error("Invalid cache dir: ", cfg.Storage.CacheDir)
				return
			}
		}
		currentDir, err := os.Getwd()
		if err != nil {
			log.Error(err)
			return
		}
		cachePath := filepath.Join(currentDir, cfg.Storage.CacheDir)
		cachePath, err = filepath.Abs(cachePath)
		if err != nil {
			log.Error(err)
			return
		}
		log.Info("Removing cache dir: ", cachePath)
		if err := os.RemoveAll(cachePath); err != nil {
			log.Error(err)
			return
		}
	}
}
