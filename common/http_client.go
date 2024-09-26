package common

import (
	"ManyACG/config"
	"context"
	"os"
	"time"

	. "ManyACG/logger"

	"github.com/imroc/req/v3"
)

var Client *req.Client

func initHttpClient() {
	c := req.C().ImpersonateChrome().SetCommonRetryCount(2).SetTLSHandshakeTimeout(time.Second * 10)
	Client = c
}

func DownloadWithCache(ctx context.Context, url string, client *req.Client) ([]byte, error) {
	if client == nil {
		client = Client
	}
	cachePath := config.Cfg.Storage.CacheDir + "/req/" + EscapeFileName(url)
	data, err := os.ReadFile(cachePath)
	if err == nil {
		Logger.Debugf("Cache hit: %s", url)
		return data, nil
	}
	Logger.Debugf("downloading: %s", url)
	resp, err := client.R().SetContext(ctx).Get(url)
	if err != nil {
		return nil, err
	}
	data = resp.Bytes()

	go func() {
		if err := MkFile(cachePath, data); err != nil {
			Logger.Errorf("failed to save cache file: %s", err)
		} else {
			go PurgeFileAfter(cachePath, time.Duration(config.Cfg.Storage.CacheTTL)*time.Second)
		}
	}()

	return data, nil
}

func GetReqCachedFile(path string) []byte {
	cachePath := config.Cfg.Storage.CacheDir + "/req/" + EscapeFileName(path)
	data, err := os.ReadFile(cachePath)
	if err != nil {
		return nil
	}
	return data
}
