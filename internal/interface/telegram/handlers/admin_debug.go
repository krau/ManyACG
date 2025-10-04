package handlers

import (
	"encoding/json"
	"fmt"
	"html"

	"github.com/krau/ManyACG/internal/interface/telegram/handlers/utils"
	"github.com/krau/ManyACG/service"
	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegohandler"
	"github.com/samber/oops"
)

func DumpArtworkInfo(ctx *telegohandler.Context, message telego.Message) error {
	serv := service.FromContext(ctx)
	userAdmin, err := serv.GetAdminByTelegramID(ctx, message.From.ID)
	if err != nil {
		// log.Errorf("获取管理员信息失败: %s", err)
		return err
		// telegoutil.ReplyMessage(ctx, ctx.Bot(), message, "获取管理员信息失败")
	}
	if userAdmin == nil {
		// utils.ReplyMessage(ctx, ctx.Bot(), message, "你没有权限执行此操作")
		return oops.Errorf("user %d is not admin but try to use /dump", message.From.ID)
	}
	helpText := fmt.Sprintf(`
[管理员] <b>使用 /dump 命令回复一条包含作品链接的消息, 将获取作品信息并以JSON格式回复</b>
	
命令语法: %s

若不提供参数, 默认获取所有信息
			`, html.EscapeString("/dump [tags] [artist] [pictures]")) // [TODO] implement this
	if message.ReplyToMessage == nil {
		utils.ReplyMessageWithHTML(ctx, message, helpText)
		return nil
	}
	sourceURL := utils.FindSourceURLInMessage(serv, message.ReplyToMessage)
	if sourceURL == "" {
		utils.ReplyMessageWithHTML(ctx, message, "回复的消息中没有支持的链接, 命令帮助:\n"+helpText)
		return nil
	}
	artwork, err := serv.GetArtworkByURL(ctx, sourceURL)
	if err != nil {
		utils.ReplyMessageWithHTML(ctx, message, fmt.Sprintf("获取作品信息失败\n<code>%s</code>", html.EscapeString(err.Error())))
		return nil
	}
	artworkJSON, err := json.MarshalIndent(artwork, "", "  ")
	if err != nil {
		utils.ReplyMessageWithHTML(ctx, message, fmt.Sprintf("序列化作品信息失败\n<code>%s</code>", html.EscapeString(err.Error())))
		return nil
	}
	if _, err := utils.ReplyMessageWithHTML(ctx, message, "<pre>"+html.EscapeString(string(artworkJSON))+"</pre>"); err != nil {
		utils.ReplyMessageWithHTML(ctx, message, fmt.Sprintf("回复消息失败\n<code>%s</code>", html.EscapeString(err.Error())))
	}
	return nil
}
