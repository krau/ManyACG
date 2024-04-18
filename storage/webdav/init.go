package webdav

import (
	"ManyACG-Bot/config"
	. "ManyACG-Bot/logger"
	"os"

	"github.com/studio-b12/gowebdav"
)

var Client *gowebdav.Client

func initClient() {
	webdavConfig := config.Cfg.Storage.Webdav
	Client = gowebdav.NewClient(webdavConfig.URL, webdavConfig.Username, webdavConfig.Password)
	if err := Client.Connect(); err != nil {
		Logger.Panicf("connect to webdav failed: %v", err)
		os.Exit(1)
	}
}

func init() {
	if config.Cfg.Storage.Type == "webdav" {
		initClient()
	}
}
