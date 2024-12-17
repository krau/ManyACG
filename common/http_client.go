package common

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/krau/ManyACG/config"

	"io"
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
	cachePath := filepath.Join(config.Cfg.Storage.CacheDir, "req", MD5Hash(url))

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
	if resp.IsErrorState() {
		return nil, fmt.Errorf("http error: %d", resp.GetStatusCode())
	}
	data = resp.Bytes()
	go MkCache(cachePath, data, time.Duration(config.Cfg.Storage.CacheTTL)*time.Second)
	return data, nil
}

func GetBodyReader(ctx context.Context, url string, client *req.Client) (io.ReadCloser, error) {
	if client == nil {
		client = Client
	}
	cachePath := filepath.Join(config.Cfg.Storage.CacheDir, "req", MD5Hash(url))
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
	cachePath := filepath.Join(config.Cfg.Storage.CacheDir, "req", MD5Hash(url))
	data, err := os.ReadFile(cachePath)
	if err != nil {
		return nil, err
	}
	return data, nil
}
