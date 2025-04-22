package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"net/http"
	_ "net/http/pprof"

	"github.com/krau/ManyACG/api/restful"
	"github.com/krau/ManyACG/common"
	"github.com/krau/ManyACG/config"
	"github.com/krau/ManyACG/dao"
	"github.com/krau/ManyACG/fetcher"
	"github.com/krau/ManyACG/service"
	"github.com/krau/ManyACG/sources"
	"github.com/krau/ManyACG/storage"
	"github.com/krau/ManyACG/telegram"
	"github.com/krau/ManyACG/webassets"
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
	config.InitConfig()
	common.Init()
	fmt.Printf(banner, common.BuildTime, common.Version, common.Commit[:7])

	if config.Cfg.Debug {
		go func() {
			common.Logger.Info("Start pprof server")
			if err := http.ListenAndServe("localhost:39060", nil); err != nil {
				common.Logger.Fatal(err)
			}
		}()
	}

	ctx, stop := signal.NotifyContext(context.TODO(), syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	defer stop()
	common.Logger.Info("Starting...")
	dao.InitDB(ctx)
	defer func() {
		if err := dao.Client.Disconnect(ctx); err != nil {
			common.Logger.Fatal(err)
		}
	}()
	service.InitService(ctx)
	sources.InitSources(service.NewService())
	storage.InitStorage(ctx)
	if config.Cfg.Telegram.Token != "" {
		telegram.RunPolling(ctx)
	}

	go fetcher.StartScheduler(ctx)
	if config.Cfg.API.Enable {
		restful.Run(ctx)
	}
	if config.Cfg.Web.Enable {
		go func() {
			common.Logger.Info("Starting serve web...")
			sm := http.NewServeMux()
			sm.Handle("/", http.FileServer(http.FS(webassets.WebAppFS)))
			if err := http.ListenAndServe(config.Cfg.Web.Address, sm); err != nil {
				common.Logger.Fatal(err)
			}
		}()
	}

	common.Logger.Info("ManyACG is running !")

	defer common.Logger.Info("Exited.")
	<-ctx.Done()
	cleanCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := service.Cleanup(cleanCtx); err != nil {
		common.Logger.Error(err)
	}
	cleanCacheDir()
}

func cleanCacheDir() {
	if config.Cfg.Storage.CacheDir != "" && !config.Cfg.Debug {
		for _, path := range []string{"/", ".", "\\", ".."} {
			if filepath.Clean(config.Cfg.Storage.CacheDir) == path {
				common.Logger.Error("Invalid cache dir: ", config.Cfg.Storage.CacheDir)
				return
			}
		}
		currentDir, err := os.Getwd()
		if err != nil {
			common.Logger.Error(err)
			return
		}
		cachePath := filepath.Join(currentDir, config.Cfg.Storage.CacheDir)
		cachePath, err = filepath.Abs(cachePath)
		if err != nil {
			common.Logger.Error(err)
			return
		}
		common.Logger.Info("Removing cache dir: ", cachePath)
		if err := os.RemoveAll(cachePath); err != nil {
			common.Logger.Error(err)
			return
		}
	}
}
