package utils

import (
	"fmt"
	"html"
	"strings"

	"github.com/krau/ManyACG/internal/interface/telegram/metautil"
	"github.com/krau/ManyACG/internal/model/entity"
	"github.com/krau/ManyACG/service"
	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegohandler"
	"github.com/mymmrac/telego/telegoutil"
)

func ReplyMessageWithHTML(ctx *telegohandler.Context, message telego.Message, text string) (*telego.Message, error) {
	return ctx.Bot().SendMessage(ctx, telegoutil.Message(message.Chat.ChatID(), text).WithReplyParameters(
		&telego.ReplyParameters{
			MessageID: message.MessageID,
		},
	).WithParseMode(telego.ModeHTML))
}

func ReplyMessage(ctx *telegohandler.Context, message telego.Message, text string) (*telego.Message, error) {
	return ctx.Bot().SendMessage(ctx, telegoutil.Message(message.Chat.ChatID(), text).WithReplyParameters(
		&telego.ReplyParameters{
			MessageID: message.MessageID,
		},
	))
}

func FindSourceURLInMessage(serv *service.Service, message *telego.Message) string {
	if message == nil {
		return ""
	}
	text := message.Text
	text += message.Caption + " "
	for _, entity := range message.Entities {
		if entity.Type == telego.EntityTypeTextLink {
			text += entity.URL + " "
		}
	}
	for _, entity := range message.CaptionEntities {
		if entity.Type == telego.EntityTypeTextLink {
			text += entity.URL + " "
		}
	}
	return serv.FindSourceURL(text)
}

var tagCharsReplacer = strings.NewReplacer(
	":", "_",
	"：", "_",
	"-", "_",
	"（", "_",
	"）", "_",
	"「", "_",
	"」", "_",
	"*", "_",
	"?", "",
	"/", " #",
	" ", "_",
)


func ArtworkHTMLCaption(meta *metautil.MetaData, artwork entity.ArtworkLike) string {
	caption := fmt.Sprintf("<a href=\"%s\"><b>%s</b></a> / <b>%s</b>", artwork.GetSourceURL(), html.EscapeString(artwork.GetTitle()), html.EscapeString(artwork.GetArtistName()))
	if artwork.GetDescription() != "" {
		desc := artwork.GetDescription()
		if len(desc) > 500 {
			var n, i int
			for i = range desc {
				if n >= 500 {
					break
				}
				n++
			}
			desc = desc[:i] + "..."
		}
		caption += fmt.Sprintf("\n<blockquote expandable=true>%s</blockquote>", html.EscapeString(desc))
	}
	tags := ""
	for _, tag := range artwork.GetTags() {
		if len(tags)+len(tag) > 200 {
			break
		}
		tag = tagCharsReplacer.Replace(tag)
		tags += "#" + strings.TrimSpace(html.EscapeString(tag)) + " "
	}
	caption += fmt.Sprintf("\n<blockquote expandable=true>%s</blockquote>\n", tags)
	// posted := meta.ChannelChatID.Username != ""
	// if posted {
	// 	caption += html.EscapeString(meta.ChannelChatID.Username)
	// }
	// if  config.Cfg.API.SiteURL != "" {
	// 	if posted {
	// 		caption += " | "
	// 	}
	// 	caption += fmt.Sprintf("<a href=\"%s/artwork/%s\">在网站查看</a>", config.Cfg.API.SiteURL, artwork.ID)
	// }
	return caption
}
