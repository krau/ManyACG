package common

import (
	"context"
	"os"
	"path/filepath"
	"time"

	"github.com/krau/ManyACG/config"

	. "github.com/krau/ManyACG/logger"

	"sync"

	"github.com/imroc/req/v3"
)

var Client *req.Client

func initHttpClient() {
	c := req.C().ImpersonateChrome().SetCommonRetryCount(2).SetTLSHandshakeTimeout(time.Second * 10)
	Client = c
}

var cacheLocks sync.Map

func DownloadWithCache(ctx context.Context, url string, client *req.Client) ([]byte, error) {
	if client == nil {
		client = Client
	}
	cachePath := filepath.Join(config.Cfg.Storage.CacheDir, "req", EscapeFileName(url))

	data, err := os.ReadFile(cachePath)
	if err == nil {
		return data, nil
	}

	lock, _ := cacheLocks.LoadOrStore(url, &sync.Mutex{})
	lock.(*sync.Mutex).Lock()
	defer func() {
		lock.(*sync.Mutex).Unlock()
		cacheLocks.Delete(url)
	}()

	data, err = os.ReadFile(cachePath)
	if err == nil {
		return data, nil
	}

	Logger.Debugf("downloading: %s", url)
	resp, err := client.R().SetContext(ctx).Get(url)
	if err != nil {
		return nil, err
	}
	data = resp.Bytes()
	go MkCache(cachePath, data, time.Duration(config.Cfg.Storage.CacheTTL)*time.Second)
	return data, nil
}

func GetReqCachedFile(path string) []byte {
	cachePath := filepath.Join(config.Cfg.Storage.CacheDir, "req", EscapeFileName(path))
	data, err := os.ReadFile(cachePath)
	if err != nil {
		return nil
	}
	return data
}
