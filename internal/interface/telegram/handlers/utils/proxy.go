package utils

import (
	"fmt"
	"net/url"

	"github.com/krau/ManyACG/internal/infra/config/runtimecfg"
	"github.com/krau/ManyACG/internal/pkg/mediatool"
)

// TelegramPhotoProxyURL 返回经图片代理缩放加边后的 photo URL.
//
// 代理按 Telegram 限制(长边不超上限, 宽高比超限时加白边)输出
func TelegramPhotoProxyURL(original string, width, height uint) string {
	base := runtimecfg.Get().Wsrv.URL
	if base == "" || original == "" || width == 0 || height == 0 {
		return ""
	}
	canvasWidth, canvasHeight := mediatool.TelegramPhotoSize(int(width), int(height))
	if canvasWidth <= 0 || canvasHeight <= 0 {
		return ""
	}
	return fmt.Sprintf("%s/?url=%s&w=%d&h=%d&fit=contain&background=%%23ffffff&we&output=jpg",
		base, url.QueryEscape(original), canvasWidth, canvasHeight)
}
