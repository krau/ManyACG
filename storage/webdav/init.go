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
		Logger.Fatalf("Failed to connect to webdav server: %v", err)
		os.Exit(1)
	}
}

func init() {
	if config.Cfg.Storage.Type == "webdav" {
		initClient()
		if Client == nil {
			Logger.Fatal("Failed to initialize webdav client")
			os.Exit(1)
		}
	}
}
