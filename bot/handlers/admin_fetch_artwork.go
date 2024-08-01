package handlers

import (
	"ManyACG/config"
	"ManyACG/fetcher"
	"ManyACG/telegram"
	"ManyACG/types"
	"context"

	"github.com/mymmrac/telego"
)

func FetchArtwork(ctx context.Context, bot *telego.Bot, message telego.Message) {
	if !CheckPermissionInGroup(ctx, message, types.PermissionFetchArtwork) {
		telegram.ReplyMessage(bot, message, "你没有拉取作品的权限")
		return
	}
	go fetcher.FetchOnce(context.TODO(), config.Cfg.Fetcher.Limit)
	telegram.ReplyMessage(bot, message, "开始拉取作品了")
}
