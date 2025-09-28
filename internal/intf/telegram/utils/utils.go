package utils

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/krau/ManyACG/internal/common"
	"github.com/krau/ManyACG/internal/common/errs"
	"github.com/krau/ManyACG/internal/infra/config"
	sources "github.com/krau/ManyACG/internal/infra/source"
	"github.com/krau/ManyACG/internal/pkg/imgtool"
	"github.com/krau/ManyACG/pkg/strutil"
	"github.com/krau/ManyACG/types"

	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegoutil"
)

func GetMessageIDs(messages []telego.Message) []int {
	ids := make([]int, len(messages))
	for i, message := range messages {
		ids[i] = message.MessageID
	}
	return ids
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

/*
	检查目标消息所回复的消息是否为一个频道的消息

如果是，返回 messageOriginChannel 和 true
*/
func GetReplyToMessageOriginChannel(message telego.Message) (*telego.MessageOriginChannel, bool) {
	if message.ReplyToMessage == nil {
		return nil, false
	}
	if message.ReplyToMessage.Photo == nil || message.ReplyToMessage.ForwardOrigin == nil {
		return nil, false
	}
	messageOriginChannel := GetMssageOriginChannel(message.ReplyToMessage)
	if messageOriginChannel == nil {
		return nil, false
	}
	return messageOriginChannel, true
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

// 获取作品的 HTML 格式描述, 已转义
func GetArtworkHTMLCaption(artwork *types.Artwork) string {
	caption := fmt.Sprintf("<a href=\"%s\"><b>%s</b></a> / <b>%s</b>", artwork.SourceURL, strutil.EscapeHTML(artwork.Title), strutil.EscapeHTML(artwork.Artist.Name))
	if artwork.Description != "" {
		desc := artwork.Description
		if len(artwork.Description) > 500 {
			var n, i int
			for i = range desc {
				if n >= 500 {
					break
				}
				n++
			}
			desc = desc[:i] + "..."
		}
		caption += fmt.Sprintf("\n<blockquote expandable=true>%s</blockquote>", strutil.EscapeHTML(desc))
	}
	tags := ""
	for _, tag := range artwork.Tags {
		if len(tags)+len(tag) > 200 {
			break
		}
		tag = tagCharsReplacer.Replace(tag)
		tags += "#" + strings.TrimSpace(strutil.EscapeHTML(tag)) + " "
	}
	caption += fmt.Sprintf("\n<blockquote expandable=true>%s</blockquote>\n", tags)
	posted := ChannelChatID.Username != "" && artwork.ID != ""
	if posted {
		caption += strutil.EscapeHTML(ChannelChatID.Username)
	}
	if artwork.ID != "" && config.Get().API.SiteURL != "" {
		if posted {
			caption += " | "
		}
		caption += fmt.Sprintf("<a href=\"%s/artwork/%s\">在网站查看</a>", config.Get().API.SiteURL, artwork.ID)
	}
	return caption
}

func ReplyMessage(ctx context.Context, bot *telego.Bot, message telego.Message, text string) (*telego.Message, error) {
	return bot.SendMessage(ctx, telegoutil.Message(message.Chat.ChatID(), text).
		WithReplyParameters(&telego.ReplyParameters{
			MessageID: message.MessageID,
		}))
}

func ReplyMessageWithMarkdown(ctx context.Context, bot *telego.Bot, message telego.Message, text string) (*telego.Message, error) {
	return bot.SendMessage(ctx, telegoutil.Message(message.Chat.ChatID(), text).
		WithReplyParameters(&telego.ReplyParameters{
			MessageID: message.MessageID,
		}).WithParseMode(telego.ModeMarkdownV2))
}

func ReplyMessageWithHTML(ctx context.Context, bot *telego.Bot, message telego.Message, text string) (*telego.Message, error) {
	return bot.SendMessage(ctx, telegoutil.Message(message.Chat.ChatID(), text).
		WithReplyParameters(&telego.ReplyParameters{
			MessageID: message.MessageID,
		}).WithParseMode(telego.ModeHTML))
}

func GetArtworkPostMessageURL(messageID int, channelChatID telego.ChatID) string {
	if channelChatID.Username != "" {
		return fmt.Sprintf("https://t.me/%s/%d", strings.TrimPrefix(channelChatID.String(), "@"), messageID)
	}
	return fmt.Sprintf("https://t.me/c/%s/%d", strings.TrimPrefix(channelChatID.String(), "-100"), messageID)
}

func GetDeepLink(botUsername, command string, args ...string) string {
	return fmt.Sprintf("https://t.me/%s/?start=%s_%s", botUsername, command, strings.Join(args, "_"))
}

func GetPostedPictureReplyMarkup(artwork *types.Artwork, index uint, channelChatID telego.ChatID, botUsername string) *telego.InlineKeyboardMarkup {
	return telegoutil.InlineKeyboard(
		GetPostedPictureInlineKeyboardButton(artwork, index, channelChatID, botUsername),
	)
}

func GetPostedPictureInlineKeyboardButton(artwork *types.Artwork, index uint, channelChatID telego.ChatID, botUsername string) []telego.InlineKeyboardButton {
	if index >= uint(len(artwork.Pictures)) {
		common.Logger.Fatalf("图片索引越界: %d", index)
		return nil
	}
	if (channelChatID.ID == 0 && channelChatID.Username == "") || (artwork.Pictures[index].TelegramInfo == nil || artwork.Pictures[index].TelegramInfo.MessageID == 0) {
		return []telego.InlineKeyboardButton{
			telegoutil.InlineKeyboardButton("详情").WithURL(config.Get().API.SiteURL + "/artwork/" + artwork.ID),
			telegoutil.InlineKeyboardButton("原图").WithURL(GetDeepLink(botUsername, "file", artwork.Pictures[index].ID)),
		}
	}
	return []telego.InlineKeyboardButton{
		telegoutil.InlineKeyboardButton("详情").WithURL(GetArtworkPostMessageURL(artwork.Pictures[index].TelegramInfo.MessageID, channelChatID)),
		telegoutil.InlineKeyboardButton("原图").WithURL(GetDeepLink(botUsername, "file", artwork.Pictures[index].ID)),
	}
}

func GetPostedArtworkInlineKeyboardButton(artwork *types.Artwork, channelChatID telego.ChatID, botUsername string) []telego.InlineKeyboardButton {
	detailsURL := config.Get().API.SiteURL + "/artwork/" + artwork.ID
	hasValidTelegramInfo := channelChatID.ID != 0 || channelChatID.Username != ""
	if hasValidTelegramInfo && artwork.Pictures[0].TelegramInfo != nil && artwork.Pictures[0].TelegramInfo.MessageID != 0 {
		detailsURL = GetArtworkPostMessageURL(artwork.Pictures[0].TelegramInfo.MessageID, channelChatID)
	}
	return []telego.InlineKeyboardButton{
		telegoutil.InlineKeyboardButton("详情").WithURL(detailsURL),
		telegoutil.InlineKeyboardButton("原图").WithURL(GetDeepLink(botUsername, "files", artwork.ID)),
	}
}

func GetMessagePhotoFile(ctx context.Context, bot *telego.Bot, message *telego.Message) ([]byte, error) {
	fileID := ""
	if message.Photo != nil {
		fileID = message.Photo[len(message.Photo)-1].FileID
	}
	if message.Document != nil && strings.HasPrefix(message.Document.MimeType, "image/") {
		if message.Document.FileSize > 20*1024*1024 {
			return nil, errs.ErrFileTooLarge
		}
		fileID = message.Document.FileID
	}
	if fileID == "" {
		return nil, errs.ErrNoPhotoInMessage
	}
	tgFile, err := bot.GetFile(ctx,
		&telego.GetFileParams{FileID: fileID},
	)
	if err != nil {
		return nil, err
	}
	return telegoutil.DownloadFile(bot.FileDownloadURL(tgFile.FilePath))
}

func FindSourceURLForMessage(message *telego.Message) string {
	if message == nil {
		common.Logger.Warn("消息为空")
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
	return sources.FindSourceURL(text)
}

// func SendPictureFileByID(ctx context.Context, bot *telego.Bot, message telego.Message, channelChatID telego.ChatID, pictureID string) (*telego.Message, error) {
// 	objectID, err := primitive.ObjectIDFromHex(pictureID)
// 	if err != nil {
// 		return nil, err
// 	}
// 	picture, err := service.GetPictureByID(ctx, objectID)
// 	if err != nil {
// 		return nil, err
// 	}
// 	var file telego.InputFile
// 	if picture.TelegramInfo != nil && picture.TelegramInfo.DocumentFileID != "" {
// 		file = telegoutil.FileFromID(picture.TelegramInfo.DocumentFileID)
// 	} else {
// 		go ReplyMessage(ctx, bot, message, "正在下载原图，请稍等~")
// 		data, err := storage.GetFile(ctx, picture.StorageInfo.Original)
// 		if err != nil {
// 			data, err = common.DownloadWithCache(ctx, picture.Original, nil)
// 			if err != nil {
// 				return nil, err
// 			}
// 		}
// 		file = telegoutil.File(telegoutil.NameReader(bytes.NewReader(data), picture.GetFileName()))
// 	}
// 	document := telegoutil.Document(message.Chat.ChatID(), file).
// 		WithReplyParameters(&telego.ReplyParameters{
// 			MessageID: message.MessageID,
// 		}).WithDisableContentTypeDetection()
// 	if IsChannelAvailable && picture.TelegramInfo != nil && picture.TelegramInfo.MessageID != 0 {
// 		document.WithReplyMarkup(telegoutil.InlineKeyboard([]telego.InlineKeyboardButton{
// 			telegoutil.InlineKeyboardButton("详情").WithURL(GetArtworkPostMessageURL(picture.TelegramInfo.MessageID, channelChatID)),
// 		}))
// 	} else {
// 		artworkID, err := primitive.ObjectIDFromHex(picture.ArtworkID)
// 		if err == nil {
// 			artwork, err := service.GetArtworkByID(ctx, artworkID)
// 			if err == nil {
// 				document.WithReplyMarkup(telegoutil.InlineKeyboard([]telego.InlineKeyboardButton{
// 					telegoutil.InlineKeyboardButton("详情").WithURL(artwork.SourceURL),
// 				}))
// 			} else {
// 				common.Logger.Warnf("获取作品信息失败: %s", err)
// 			}
// 		} else {
// 			common.Logger.Warnf("创建 ObjectID 失败: %s", err)
// 		}
// 	}
// 	documentMessage, err := bot.SendDocument(ctx, document)
// 	if err != nil {
// 		return nil, err
// 	}
// 	if documentMessage != nil {
// 		if documentMessage.Document != nil {
// 			if picture.TelegramInfo == nil {
// 				picture.TelegramInfo = &types.TelegramInfo{}
// 			}
// 			picture.TelegramInfo.DocumentFileID = documentMessage.Document.FileID
// 			if service.UpdatePictureTelegramInfo(ctx, picture, picture.TelegramInfo) != nil {
// 				common.Logger.Warnf("更新图片信息失败: %s", err)
// 			}
// 		}
// 	}
// 	return documentMessage, nil
// }

// GetPicturePreviewInputFile 获取图片预览的 InputFile
func GetPicturePreviewInputFile(ctx context.Context, picture *types.Picture) (*telego.InputFile, error) {
	if picture.TelegramInfo != nil && picture.TelegramInfo.PhotoFileID != "" {
		inputFile := telegoutil.FileFromID(picture.TelegramInfo.PhotoFileID)
		return &inputFile, nil
	}
	cacheFile, err := common.GetReqCachedFile(picture.Original)
	if err == nil {
		fileBytes, err := imgtool.CompressImageForTelegram(cacheFile)
		if err != nil {
			return nil, err
		}
		inputFile := telegoutil.File(telegoutil.NameReader(bytes.NewReader(fileBytes), picture.GetFileName()))
		return &inputFile, nil
	}
	return nil, errs.ErrNoAvailableFile
}

func ParseCommandBy(text string, splitChar, quoteChar string) (string, string, []string) {
	cmd, username, payload := telegoutil.ParseCommandPayload(text)

	if payload == "" {
		return cmd, username, []string{}
	}

	var args []string
	var currentArg strings.Builder
	inQuote := false

	for _, char := range payload {
		strChar := string(char)

		if strChar == quoteChar {
			inQuote = !inQuote
			continue
		}

		if strChar == splitChar && !inQuote {
			if currentArg.Len() > 0 {
				args = append(args, currentArg.String())
				currentArg.Reset()
			}
			continue
		}

		currentArg.WriteString(strChar)
	}

	if currentArg.Len() > 0 {
		args = append(args, currentArg.String())
	}

	return cmd, username, args
}
