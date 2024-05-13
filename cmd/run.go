package cmd

import (
	"ManyACG/api/restful"
	"ManyACG/bot"
	"ManyACG/config"
	"ManyACG/dao"
	"ManyACG/fetcher"
	"ManyACG/logger"
	"ManyACG/sources"
	"context"
	"os"
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
	sources.InitSources()
	go bot.RunPolling()
	go fetcher.StartScheduler(context.TODO())
	if config.Cfg.API.Enable {
		go restful.Run()
	}
	select {}
}
