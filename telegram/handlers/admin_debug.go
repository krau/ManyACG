package handlers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/krau/ManyACG/common"
	"github.com/krau/ManyACG/service"
	"github.com/krau/ManyACG/sources"
	"github.com/krau/ManyACG/telegram/utils"
	"github.com/mymmrac/telego"
)

func DumpArtworkInfo(ctx context.Context, bot *telego.Bot, message telego.Message) {
	userAdmin, err := service.GetAdminByUserID(ctx, message.From.ID)
	if err != nil {
		common.Logger.Errorf("获取管理员信息失败: %s", err)
		utils.ReplyMessage(bot, message, "获取管理员信息失败")
		return
	}
	if userAdmin == nil {
		utils.ReplyMessage(bot, message, "你没有权限执行此操作")
		return
	}

	var sourceURL string
	if message.ReplyToMessage != nil {
		sourceURL = utils.FindSourceURLForMessage(message.ReplyToMessage)
	} else {
		sourceURL = sources.FindSourceURL(message.Text)
	}
	if sourceURL == "" {
		utils.ReplyMessage(bot, message, "请回复一条消息, 或者指定作品链接")
		return
	}

	artwork, err := service.GetArtworkByURL(ctx, sourceURL)
	if err != nil {
		common.Logger.Errorf("获取作品信息失败: %s", err)
		utils.ReplyMessageWithHTML(bot, message, fmt.Sprintf("获取作品信息失败\n<code>%s</code>", common.EscapeHTML(err.Error())))
		return
	}
	artworkJSON, err := json.MarshalIndent(artwork, "", "  ")
	if err != nil {
		common.Logger.Errorf("序列化作品信息失败: %s", err)
		utils.ReplyMessageWithHTML(bot, message, fmt.Sprintf("序列化作品信息失败\n<code>%s</code>", common.EscapeHTML(err.Error())))
		return
	}
	utils.ReplyMessageWithHTML(bot, message, "<pre>"+common.EscapeHTML(string(artworkJSON))+"</pre>")
}
