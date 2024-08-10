package handlers

import (
	. "ManyACG/logger"
	"ManyACG/service"
	"ManyACG/telegram"
	"context"
	"fmt"

	"github.com/mymmrac/telego"
)

func GetStats(ctx context.Context, bot *telego.Bot, message telego.Message) {
	stats, err := service.GetDatabaseStats(ctx)
	if err != nil {
		Logger.Errorf("获取统计信息失败: %s", err)
		telegram.ReplyMessage(bot, message, "获取统计信息失败: "+err.Error())
		return
	}
	text := fmt.Sprintf(
		"关于数据库可以公开的情报:\n\n总图片数: %d\n总标签数: %d\n总画师数: %d\n总作品数: %d",
		stats.TotalPictures, stats.TotalTags, stats.TotalArtists, stats.TotalArtworks, //stats.LastArtworkUpdate.Format("2006-01-02 15:04:05"),
	)
	telegram.ReplyMessage(bot, message, text)
}
