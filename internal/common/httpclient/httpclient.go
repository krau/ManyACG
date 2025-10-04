package httpclient

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/imroc/req/v3"
	"github.com/krau/ManyACG/internal/infra/config/runtimecfg"
	"github.com/krau/ManyACG/pkg/osutil"
	"github.com/krau/ManyACG/pkg/strutil"
)

var (
	defaultClient *req.Client
	once          sync.Once
	cacheLocks    sync.Map
)

func initDefaultClient() {
	c := req.C().ImpersonateChrome().
		SetCommonRetryCount(2).
		SetTLSHandshakeTimeout(time.Second * 10).
		SetTimeout(time.Minute * 2)
	defaultClient = c
	if proxyUrl := runtimecfg.Get().HttpClient.Proxy; proxyUrl != "" {
		defaultClient.SetProxyURL(proxyUrl)
	}
}

func getCachePath(url string) string {
	return filepath.Join(runtimecfg.Get().Storage.CacheDir, "req", strutil.MD5Hash(url))
}

func DownloadWithCache(ctx context.Context, url string, client *req.Client) ([]byte, error) {
	once.Do(initDefaultClient)
	if client == nil {
		client = defaultClient
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

	resp, err := client.R().SetContext(ctx).Get(url)
	if err != nil {
		return nil, err
	}
	if resp.IsErrorState() {
		return nil, fmt.Errorf("http error: %d", resp.GetStatusCode())
	}
	data = resp.Bytes()
	osutil.MkCache(cachePath, data, time.Duration(runtimecfg.Get().Storage.CacheTTL)*time.Second)
	return data, nil
}

func GetBodyReader(ctx context.Context, url string, client *req.Client) (io.ReadCloser, error) {
	once.Do(initDefaultClient)
	if client == nil {
		client = defaultClient
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
