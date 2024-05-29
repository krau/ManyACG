package common

import (
	"time"

	"github.com/imroc/req/v3"
)

var Client *req.Client

func init() {
	c := req.C().ImpersonateChrome().SetCommonRetryCount(2).SetUserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/117.0.0.0 Safari/537.36")
	c.TLSHandshakeTimeout = time.Second * 10
	Client = c
}

func DownloadFromURL(url string) ([]byte, error) {
	resp, err := Client.R().Get(url)
	if err != nil {
		return nil, err
	}
	return resp.Bytes(), nil
}
