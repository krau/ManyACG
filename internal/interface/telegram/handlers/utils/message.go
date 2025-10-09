package utils

import (
	"fmt"
	"html"
	"strings"

	"github.com/duke-git/lancet/v2/strutil"
	"github.com/krau/ManyACG/internal/interface/telegram/metautil"
	"github.com/krau/ManyACG/internal/model/entity"
	"github.com/krau/ManyACG/internal/service"
	"github.com/krau/ManyACG/internal/shared"
	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegohandler"
	"github.com/mymmrac/telego/telegoutil"
	"github.com/samber/oops"
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
	var sb strings.Builder
	sb.WriteString(message.Text)
	sb.WriteString(" ")
	sb.WriteString(message.Caption)
	for _, entity := range message.Entities {
		if entity.Type == telego.EntityTypeTextLink {
			sb.WriteString(entity.URL)
			sb.WriteString(" ")
		}
	}
	for _, entity := range message.CaptionEntities {
		if entity.Type == telego.EntityTypeTextLink {
			sb.WriteString(entity.URL)
			sb.WriteString(" ")
		}
	}
	return serv.FindSourceURL(sb.String())
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

func ArtworkHTMLCaption(meta *metautil.MetaData, artwork shared.ArtworkLike) string {
	tmpl := "<a href='%s'><b>%s</b></a> / <b>%s</b>"
	sourceUrl := artwork.GetSourceURL()
	title := html.EscapeString(artwork.GetTitle())
	artistName := html.EscapeString(artwork.GetArtistName())
	description := html.EscapeString(strutil.Ellipsis(artwork.GetDescription(), 500))

	tags := ""
	for _, tag := range artwork.GetTags() {
		if len(tags)+len(tag) > 200 {
			break
		}
		tag = tagCharsReplacer.Replace(tag)
		tag = strings.Trim(tag, "_")
		tags += "#" + strings.TrimSpace(html.EscapeString(tag)) + " "
	}

	args := []any{sourceUrl, title, artistName}

	if description != "" {
		tmpl += "\n<blockquote expandable=true>%s</blockquote>"
		args = append(args, description)
	}
	if tags != "" {
		tmpl += "\n<blockquote expandable=true>%s</blockquote>"
		args = append(args, tags)
	}

	caption := fmt.Sprintf(tmpl, args...)
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
	var detailsURL string
	if meta.SiteURL() != "" {
		detailsURL = fmt.Sprintf("%s/artwork/%s", meta.SiteURL(), artwork.ID.Hex())
	} else {
		detailsURL = artwork.SourceURL
	}
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
	if len(message.Photo) == 0 {
		return nil, oops.New("not a photo message")
	}
	size := message.Photo[len(message.Photo)-1]
	tfile, err := ctx.Bot().GetFile(ctx, &telego.GetFileParams{FileID: size.FileID})
	if err != nil {
		return nil, oops.Wrapf(err, "get file info failed")
	}
	dlUrl := ctx.Bot().FileDownloadURL(tfile.FilePath)
	// file, clean, err := httpclient.DownloadWithCache(ctx, dlUrl, nil)
	// if err != nil {
	// 	return nil, oops.Wrapf(err, "download file failed")
	// }
	// defer clean()
	// defer file.Close()
	// data, err := io.ReadAll(file)
	// if err != nil {
	// 	return nil, oops.Wrapf(err, "read file failed")
	// }
	// return data, nil
	return telegoutil.DownloadFile(dlUrl)
}
