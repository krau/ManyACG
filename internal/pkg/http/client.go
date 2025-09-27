package http

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/imroc/req/v3"
	"github.com/krau/ManyACG/internal/infra/config"
)

var Client *req.Client

func initHttpClient() {
	c := req.C().ImpersonateChrome().SetCommonRetryCount(2).SetTLSHandshakeTimeout(time.Second * 10).SetTimeout(time.Minute * 2)
	Client = c
	if config.Cfg.Source.Proxy != "" {
		Client.SetProxyURL(config.Cfg.Source.Proxy)
	}
}

var cacheLocks sync.Map

func getCachePath(url string) string {
	return filepath.Join(config.Cfg.Storage.CacheDir, "req", MD5Hash(url))
}

func DownloadWithCache(ctx context.Context, url string, client *req.Client) ([]byte, error) {
	if client == nil {
		client = Client
	}
	cachePath := getCachePath(url)

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

	Logger.Debugf("downloading: %s", url)
	resp, err := client.R().SetContext(ctx).Get(url)
	if err != nil {
		return nil, err
	}
	if resp.IsErrorState() {
		return nil, fmt.Errorf("http error: %d", resp.GetStatusCode())
	}
	data = resp.Bytes()
	MkCache(cachePath, data, time.Duration(config.Cfg.Storage.CacheTTL)*time.Second)
	return data, nil
}

func GetBodyReader(ctx context.Context, url string, client *req.Client) (io.ReadCloser, error) {
	if client == nil {
		client = Client
	}
	cachePath := getCachePath(url)
	if file, err := os.Open(cachePath); err == nil {
		return file, nil
	}

	lock, _ := cacheLocks.LoadOrStore(url, &sync.Mutex{})
	lock.(*sync.Mutex).Lock()
	defer func() {
		lock.(*sync.Mutex).Unlock()
		cacheLocks.Delete(url)
	}()

	if file, err := os.Open(cachePath); err == nil {
		return file, nil
	}
	Logger.Debugf("getting: %s", url)
	resp, err := client.R().SetContext(ctx).Get(url)
	if err != nil {
		return nil, err
	}
	if resp.IsErrorState() {
		return nil, fmt.Errorf("http error: %d", resp.GetStatusCode())
	}
	return resp.Body, nil
}

func GetReqCachedFile(url string) ([]byte, error) {
	cachePath := getCachePath(url)
	data, err := os.ReadFile(cachePath)
	if err != nil {
		return nil, err
	}
	return data, nil
}
