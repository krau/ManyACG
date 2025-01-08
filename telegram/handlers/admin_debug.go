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
	"github.com/mymmrac/telego/telegoutil"
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

	_, _, args := telegoutil.ParseCommand(message.Text)
	if len(args) == 0 && message.ReplyToMessage == nil {
		utils.ReplyMessage(bot, message, "请提供作品链接, 或回复一条消息")
		return
	}
	var sourceURL string
	if message.ReplyToMessage != nil {
		sourceURL = utils.FindSourceURLForMessage(message.ReplyToMessage)
		if sourceURL == "" {
			if len(args) == 0 {
				utils.ReplyMessage(bot, message, "不支持的链接")
				return
			}
		}
	}
	if len(args) > 0 {
		sourceURL = sources.FindSourceURL(args[0])
	}
	if sourceURL == "" {
		utils.ReplyMessage(bot, message, "不支持的链接")
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
