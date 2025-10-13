package httpclient

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/imroc/req/v3"
	"github.com/krau/ManyACG/internal/infra/config/runtimecfg"
	"github.com/krau/ManyACG/pkg/log"
	"github.com/krau/ManyACG/pkg/osutil"
	"github.com/krau/ManyACG/pkg/strutil"
	"golang.org/x/sync/singleflight"
)

var (
	defaultClient *req.Client
	once          sync.Once
	dlGroup       singleflight.Group
)

func initDefaultClient() {
	c := req.C().
		ImpersonateChrome().
		SetCommonRetryCount(2).
		SetLogger(log.Default()).
		EnableDebugLog()
	defaultClient = c
	if proxyUrl := runtimecfg.Get().HttpClient.Proxy; proxyUrl != "" {
		defaultClient.SetProxyURL(proxyUrl)
	}
}

func getCachePath(url string) string {
	ext, _ := strutil.GetFileExtFromURL(url)
	return filepath.Join(runtimecfg.Get().Storage.CacheDir, "req", strutil.MD5Hash(url)+ext)
}

// DownloadWithCache downloads a file with caching. If the file is already cached, it returns the cached file.
func DownloadWithCache(ctx context.Context, url string, client *req.Client) (
	*osutil.File,
	error,
) {
	once.Do(initDefaultClient)
	if client == nil {
		client = defaultClient
	}
	cachePath := getCachePath(url)
	if fi, err := os.Stat(cachePath); err == nil && !fi.IsDir() {
		return osutil.OpenCache(cachePath)
	} else if err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	ch := dlGroup.DoChan(cachePath, func() (any, error) {
		if fi, err := os.Stat(cachePath); err == nil && !fi.IsDir() {
			return nil, nil
		}
		resp, err := client.R().
			SetContext(ctx).
			SetOutputFile(cachePath).
			Get(url)
		if err != nil {
			return nil, err
		}
		if resp.IsErrorState() {
			return nil, fmt.Errorf("http error: %d", resp.GetStatusCode())
		}
		return nil, nil
	})
	select {
	case r := <-ch:
		if r.Err != nil {
			return nil, r.Err
		}
		return osutil.OpenCache(cachePath)
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}
