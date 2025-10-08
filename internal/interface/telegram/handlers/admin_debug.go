package handlers

import (
	"encoding/json"
	"fmt"
	"html"

	"github.com/krau/ManyACG/internal/interface/telegram/handlers/utils"
	"github.com/krau/ManyACG/service"
	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegohandler"
	"github.com/mymmrac/telego/telegoutil"
	"github.com/samber/oops"
)

func DumpArtworkInfo(ctx *telegohandler.Context, message telego.Message) error {
	serv := service.FromContext(ctx)
	_, err := serv.GetAdminByTelegramID(ctx, message.From.ID)
	if err != nil {
		return oops.Errorf("get admin by telegram id %d failed: %w", message.From.ID, err)
	}
	helpText := "[管理员] <b>使用 /dump 命令回复一条包含作品链接的消息, 将获取作品信息并以JSON格式回复</b>"
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
	_, err = ctx.Bot().SendDocument(ctx, telegoutil.Document(message.Chat.ChatID(),
		telegoutil.FileFromBytes(artworkJSON, fmt.Sprintf("artwork_%s.json", artwork.ID.String()))).
		WithReplyParameters(&telego.ReplyParameters{
			MessageID: message.MessageID,
		}))
	if err != nil {
		return oops.Errorf("send artwork json document failed: %w", err)
	}
	return nil
}
