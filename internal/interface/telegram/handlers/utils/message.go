package utils

import (
	"fmt"
	"html"
	"strings"

	"github.com/krau/ManyACG/internal/infra/config/runtimecfg"
	"github.com/krau/ManyACG/internal/interface/telegram/metautil"
	"github.com/krau/ManyACG/internal/model/entity"
	"github.com/krau/ManyACG/pkg/objectuuid"
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
	// [TODO] implement channel signature
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

func GetMssageOriginChannel(message *telego.Message) *telego.MessageOriginChannel {
	if message.ForwardOrigin == nil {
		return nil
	}
	if message.ForwardOrigin.OriginType() == telego.OriginTypeChannel {
		return message.ForwardOrigin.(*telego.MessageOriginChannel)
	} else {
		return nil
	}
}

func GetPostedArtworkInlineKeyboardButton(artwork *entity.Artwork, meta *metautil.MetaData) []telego.InlineKeyboardButton {
	detailsURL := runtimecfg.Get().API.SiteURL + "/artwork/" + artwork.ID.Hex() // [TODO] refactor this
	hasValidTelegramInfo := meta.ChannelChatID().ID != 0 || meta.ChannelChatID().Username != ""
	if hasValidTelegramInfo && artwork.Pictures[0].TelegramInfo.Data().MessageID != 0 {
		detailsURL = meta.ChannelMessageURL(artwork.Pictures[0].TelegramInfo.Data().MessageID)
	}
	return []telego.InlineKeyboardButton{
		telegoutil.InlineKeyboardButton("详情").WithURL(detailsURL),
		telegoutil.InlineKeyboardButton("原图").WithURL(meta.BotDeepLink("files", artwork.ID.Hex())),
	}
}

func GetMessagePhotoFile(ctx *telegohandler.Context, message *telego.Message) ([]byte, error) {
	panic("unimplemented")
}

func SendPictureFileByID(ctx *telegohandler.Context, meta *metautil.MetaData, id objectuuid.ObjectUUID) (telego.Message, error) {
	panic("unimplemented")
}

func GetPostedPictureInlineKeyboardButton(artwork *entity.Artwork, picIndex uint, meta *metautil.MetaData) []telego.InlineKeyboardButton {
	panic("unimplemented")
}
