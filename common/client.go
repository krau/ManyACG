package common

import (
	"ManyACG/config"
	"os"
	"time"

	. "ManyACG/logger"

	"github.com/imroc/req/v3"
)

var Client *req.Client

func init() {
	c := req.C().ImpersonateChrome().SetCommonRetryCount(2)
	c.TLSHandshakeTimeout = time.Second * 10
	Client = c
}

func DownloadWithCache(url string, client *req.Client) ([]byte, error) {
	if client == nil {
		client = Client
	}
	cachePath := config.Cfg.Storage.CacheDir + "/" + ReplaceFileNameInvalidChar(url)
	data, err := os.ReadFile(cachePath)
	if err == nil {
		Logger.Debugf("cache hit: %s", cachePath)
		return data, nil
	}
	resp, err := client.R().Get(url)
	if err != nil {
		return nil, err
	}
	data = resp.Bytes()
	if err := MkFile(cachePath, data); err != nil {
		Logger.Errorf("failed to save cache file: %s", err)
	} else {
		go PurgeFileAfter(cachePath, time.Duration(config.Cfg.Storage.CacheTTL)*time.Second)
	}
	return data, nil
}
