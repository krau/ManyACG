package pixiv

import (
	"ManyACG-Bot/config"
	"net/http"

	"github.com/imroc/req/v3"
)

var ReqClient *req.Client

func init() {
	ReqClient = req.C().ImpersonateChrome()
	cookies := make([]*http.Cookie, 0)
	for name, value := range config.Cfg.Source.Pixiv.Cookies {
		cookies = append(cookies, &http.Cookie{
			Name:  name,
			Value: value,
		})
	}
	ReqClient.SetCommonCookies(cookies...)
	if config.Cfg.Source.Proxy != "" {
		ReqClient.SetProxyURL(config.Cfg.Source.Proxy)
	}
}
