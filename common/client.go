package common

import (
	"ManyACG-Bot/config"
	"net/http"
	"time"

	"github.com/imroc/req/v3"
)

var Cilent *req.Client

func init() {
	c := req.C().ImpersonateChrome().SetCommonRetryCount(2).SetTimeout(time.Second * 10)
	c.TLSHandshakeTimeout = time.Second * 10
	if config.Cfg.Source.Pixiv.Cookies != "" {
		c = c.SetCommonCookies(&http.Cookie{
			Name:  "PHPSESSID",
			Value: config.Cfg.Source.Pixiv.Cookies,
		})
	}
	Cilent = c
}
