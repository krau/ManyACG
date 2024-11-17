package cmd

import (
	"context"
	"os"
	"os/signal"
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
)

func Run() {
	config.InitConfig()
	common.Init()

	if config.Cfg.Debug {
		go func() {
			common.Logger.Info("Start pprof server")
			if err := http.ListenAndServe("localhost:39060", nil); err != nil {
				common.Logger.Fatal(err)
			}
		}()
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	common.Logger.Info("Start running")
	dao.InitDB(ctx)
	defer func() {
		if err := dao.Client.Disconnect(ctx); err != nil {
			common.Logger.Fatal(err)
			os.Exit(1)
		}
	}()
	if config.Cfg.Telegram.Token != "" {
		go telegram.RunPolling()
	}
	storage.InitStorage()
	sources.InitSources()
	go fetcher.StartScheduler(context.TODO())
	if config.Cfg.API.Enable {
		go restful.Run()
	}
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit
	common.Logger.Info(sig, " Exiting...")
	if err := service.Cleanup(context.TODO()); err != nil {
		common.Logger.Error(err)
	}
	common.Logger.Info("See you next time.")
}
