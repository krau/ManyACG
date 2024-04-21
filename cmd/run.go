package cmd

import (
	"ManyACG-Bot/api/restful"
	"ManyACG-Bot/bot"
	"ManyACG-Bot/config"
	"ManyACG-Bot/dao"
	"ManyACG-Bot/fetcher"
	. "ManyACG-Bot/logger"
	"context"
	"os"
	"time"
)

func Run() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	Logger.Info("Start running")
	dao.InitDB(ctx)
	defer func() {
		if err := dao.Client.Disconnect(ctx); err != nil {
			Logger.Fatal(err)
			os.Exit(1)
		}
	}()
	go bot.RunPolling()
	go fetcher.StartScheduler(context.TODO())
	if config.Cfg.API.Enable {
		go restful.Run()
	}
	select {}
}
