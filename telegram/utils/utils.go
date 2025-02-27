package utils

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/krau/ManyACG/common"
	"github.com/krau/ManyACG/config"
	"github.com/krau/ManyACG/errs"
	"github.com/krau/ManyACG/service"
	"github.com/krau/ManyACG/sources"
	"github.com/krau/ManyACG/storage"
	"github.com/krau/ManyACG/types"

	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegoutil"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func GetMessageIDs(messages []telego.Message) []int {
	ids := make([]int, len(messages))
	for i, message := range messages {
		ids[i] = message.MessageID
	}
	return ids
}

func GetMssageOriginChannel(message telego.Message) *telego.MessageOriginChannel {
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
	检查目标消息是否为频道的作品消息

如果是，返回 messageOriginChannel 和 true
*/
func GetMessageOriginChannelArtworkPost(ctx context.Context, bot *telego.Bot, message telego.Message) (*telego.MessageOriginChannel, bool) {
	if message.ReplyToMessage == nil {
		return nil, false
	}
	if message.ReplyToMessage.Photo == nil || message.ReplyToMessage.ForwardOrigin == nil {
		return nil, false
	}
	messageOriginChannel := GetMssageOriginChannel(*message.ReplyToMessage)
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
	caption := fmt.Sprintf("<a href=\"%s\"><b>%s</b></a>", artwork.SourceURL, common.EscapeHTML(artwork.Title))
	caption += fmt.Sprintf("\nAuthor: <b>%s</b>", common.EscapeHTML(artwork.Artist.Name))
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
		caption += fmt.Sprintf("\n<blockquote expandable=true>%s</blockquote>", common.EscapeHTML(desc))
	}
	tags := ""
	for _, tag := range artwork.Tags {
		if len(tags)+len(tag) > 200 {
			break
		}
		tag = tagCharsReplacer.Replace(tag)
		tags += "#" + strings.TrimSpace(common.EscapeHTML(tag)) + " "
	}
	caption += fmt.Sprintf("\n<blockquote expandable=true>%s</blockquote>\n", tags)
	posted := ChannelChatID.Username != "" && artwork.ID != ""
	if posted {
		caption += common.EscapeHTML(ChannelChatID.Username)
	}
	if artwork.ID != "" && config.Cfg.API.SiteURL != "" {
		if posted {
			caption += " | "
		}
		caption += fmt.Sprintf("<a href=\"%s/artwork/%s\">在网站查看</a>", config.Cfg.API.SiteURL, artwork.ID)
	}
	return caption
}

func ReplyMessage(bot *telego.Bot, message telego.Message, text string) (*telego.Message, error) {
	return bot.SendMessage(telegoutil.Message(message.Chat.ChatID(), text).
		WithReplyParameters(&telego.ReplyParameters{
			MessageID: message.MessageID,
		}))
}

func ReplyMessageWithMarkdown(bot *telego.Bot, message telego.Message, text string) (*telego.Message, error) {
	return bot.SendMessage(telegoutil.Message(message.Chat.ChatID(), text).
		WithReplyParameters(&telego.ReplyParameters{
			MessageID: message.MessageID,
		}).WithParseMode(telego.ModeMarkdownV2))
}

func ReplyMessageWithHTML(bot *telego.Bot, message telego.Message, text string) (*telego.Message, error) {
	return bot.SendMessage(telegoutil.Message(message.Chat.ChatID(), text).
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
			telegoutil.InlineKeyboardButton("详情").WithURL(config.Cfg.API.SiteURL + "/artwork/" + artwork.ID),
			telegoutil.InlineKeyboardButton("原图").WithURL(GetDeepLink(botUsername, "file", artwork.Pictures[index].ID)),
		}
	}
	return []telego.InlineKeyboardButton{
		telegoutil.InlineKeyboardButton("详情").WithURL(GetArtworkPostMessageURL(artwork.Pictures[index].TelegramInfo.MessageID, channelChatID)),
		telegoutil.InlineKeyboardButton("原图").WithURL(GetDeepLink(botUsername, "file", artwork.Pictures[index].ID)),
	}
}

func GetMessagePhotoFile(bot *telego.Bot, message *telego.Message) ([]byte, error) {
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
	tgFile, err := bot.GetFile(
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

func SendPictureFileByID(ctx context.Context, bot *telego.Bot, message telego.Message, channelChatID telego.ChatID, pictureID string) (*telego.Message, error) {
	objectID, err := primitive.ObjectIDFromHex(pictureID)
	if err != nil {
		return nil, err
	}
	picture, err := service.GetPictureByID(ctx, objectID)
	if err != nil {
		return nil, err
	}
	var file telego.InputFile
	if picture.TelegramInfo != nil && picture.TelegramInfo.DocumentFileID != "" {
		file = telegoutil.FileFromID(picture.TelegramInfo.DocumentFileID)
	} else {
		go ReplyMessage(bot, message, "正在下载原图，请稍等~")
		data, err := storage.GetFile(ctx, picture.StorageInfo.Original)
		if err != nil {
			data, err = common.DownloadWithCache(ctx, picture.Original, nil)
			if err != nil {
				return nil, err
			}
		}
		file = telegoutil.File(telegoutil.NameReader(bytes.NewReader(data), picture.GetFileName()))
	}
	document := telegoutil.Document(message.Chat.ChatID(), file).
		WithReplyParameters(&telego.ReplyParameters{
			MessageID: message.MessageID,
		}).WithDisableContentTypeDetection()
	if IsChannelAvailable && picture.TelegramInfo != nil && picture.TelegramInfo.MessageID != 0 {
		document.WithReplyMarkup(telegoutil.InlineKeyboard([]telego.InlineKeyboardButton{
			telegoutil.InlineKeyboardButton("详情").WithURL(GetArtworkPostMessageURL(picture.TelegramInfo.MessageID, channelChatID)),
		}))
	} else {
		artworkID, err := primitive.ObjectIDFromHex(picture.ArtworkID)
		if err == nil {
			artwork, err := service.GetArtworkByID(ctx, artworkID)
			if err == nil {
				document.WithReplyMarkup(telegoutil.InlineKeyboard([]telego.InlineKeyboardButton{
					telegoutil.InlineKeyboardButton("详情").WithURL(artwork.SourceURL),
				}))
			} else {
				common.Logger.Warnf("获取作品信息失败: %s", err)
			}
		} else {
			common.Logger.Warnf("创建 ObjectID 失败: %s", err)
		}
	}
	documentMessage, err := bot.SendDocument(document)
	if err != nil {
		return nil, err
	}
	if documentMessage != nil {
		if documentMessage.Document != nil {
			if picture.TelegramInfo == nil {
				picture.TelegramInfo = &types.TelegramInfo{}
			}
			picture.TelegramInfo.DocumentFileID = documentMessage.Document.FileID
			if service.UpdatePictureTelegramInfo(ctx, picture, picture.TelegramInfo) != nil {
				common.Logger.Warnf("更新图片信息失败: %s", err)
			}
		}
	}
	return documentMessage, nil
}

func GetPicturePreviewInputFile(ctx context.Context, picture *types.Picture) (*telego.InputFile, bool, error) {
	if picture.TelegramInfo != nil && picture.TelegramInfo.PhotoFileID != "" {
		inputFile := telegoutil.FileFromID(picture.TelegramInfo.PhotoFileID)
		return &inputFile, false, nil
	}
	cacheFile, err := common.GetReqCachedFile(picture.Original)
	if err == nil {
		fileBytes, err := common.CompressImageForTelegramByFFmpegFromBytes(cacheFile)
		if err != nil {
			return nil, false, err
		}
		inputFile := telegoutil.File(telegoutil.NameReader(bytes.NewReader(fileBytes), picture.Original))
		return &inputFile, false, nil
	}
	if botPhotoFileID := service.GetEtcData(ctx, "bot_photo_file_id"); botPhotoFileID != nil {
		inputFile := telegoutil.FileFromID(botPhotoFileID.(string))
		return &inputFile, true, nil
	}
	if botPhotoFileBytes := service.GetEtcData(ctx, "bot_photo_bytes"); botPhotoFileBytes != nil {
		if data, ok := botPhotoFileBytes.(primitive.Binary); ok {
			inputFile := telegoutil.File(telegoutil.NameReader(bytes.NewReader(data.Data), picture.Original))
			return &inputFile, true, nil
		}
	}
	return nil, true, errs.ErrNoAvailableFile
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
