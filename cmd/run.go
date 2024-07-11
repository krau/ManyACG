package cmd

import (
	"ManyACG/api/restful"
	"ManyACG/bot"
	"ManyACG/config"
	"ManyACG/dao"
	"ManyACG/fetcher"
	"ManyACG/logger"
	"ManyACG/service"
	"ManyACG/sources"
	"ManyACG/storage"
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func Run() {
	config.InitConfig()
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
	go bot.RunPolling()
	go storage.InitStorage()
	go sources.InitSources()
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
