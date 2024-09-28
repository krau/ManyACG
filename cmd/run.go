package cmd

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/krau/ManyACG/api/restful"
	"github.com/krau/ManyACG/common"
	"github.com/krau/ManyACG/config"
	"github.com/krau/ManyACG/dao"
	"github.com/krau/ManyACG/fetcher"
	"github.com/krau/ManyACG/logger"
	"github.com/krau/ManyACG/service"
	"github.com/krau/ManyACG/sources"
	"github.com/krau/ManyACG/storage"
	"github.com/krau/ManyACG/telegram"
)

func Run() {
	config.InitConfig()
	common.Init()
	logger.InitLogger()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	logger.Logger.Info("Start running")
	dao.InitDB(ctx)
	defer func() {
		if err := dao.Client.Disconnect(ctx); err != nil {
			logger.Logger.Fatal(err)
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
	logger.Logger.Info(sig, " Exiting...")
	if err := service.Cleanup(context.TODO()); err != nil {
		logger.Logger.Error(err)
	}
	logger.Logger.Info("See you next time.")
}
