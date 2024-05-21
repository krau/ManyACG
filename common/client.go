package common

import (
	"time"

	"github.com/imroc/req/v3"
)

var Client *req.Client

func init() {
	c := req.C().ImpersonateChrome().SetCommonRetryCount(2).SetTimeout(time.Second * 10)
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
