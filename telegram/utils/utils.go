package utils

import (
	"bytes"
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/krau/ManyACG/common"
	"github.com/krau/ManyACG/config"
	"github.com/krau/ManyACG/errors"
	"github.com/krau/ManyACG/service"
	"github.com/krau/ManyACG/sources"
	"github.com/krau/ManyACG/storage"
	"github.com/krau/ManyACG/types"

	. "github.com/krau/ManyACG/logger"

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

func GetMssageOriginChannel(_ context.Context, _ *telego.Bot, message telego.Message) *telego.MessageOriginChannel {
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
	messageOriginChannel := GetMssageOriginChannel(ctx, bot, *message.ReplyToMessage)
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
)

func GetArtworkHTMLCaption(artwork *types.Artwork) string {
	caption := fmt.Sprintf("<a href=\"%s\"><b>%s</b></a>", artwork.SourceURL, common.EscapeHTML(artwork.Title))
	caption += "\n<b>Author:</b> " + common.EscapeHTML(artwork.Artist.Name)
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
	caption += fmt.Sprintf("\n<blockquote expandable=true>%s</blockquote>", tags)
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

func GetDeepLinkForFile(pictureID string, botUsername string) string {
	return fmt.Sprintf("https://t.me/%s/?start=file_%s", botUsername, pictureID)
}

func GetPostedPictureReplyMarkup(artwork *types.Artwork, index uint, channelChatID telego.ChatID, botUsername string) *telego.InlineKeyboardMarkup {
	return telegoutil.InlineKeyboard(
		GetPostedPictureInlineKeyboardButton(artwork, index, channelChatID, botUsername),
	)
}

func GetPostedPictureInlineKeyboardButton(artwork *types.Artwork, index uint, channelChatID telego.ChatID, botUsername string) []telego.InlineKeyboardButton {
	if index >= uint(len(artwork.Pictures)) {
		Logger.Fatalf("图片索引越界: %d", index)
		return nil
	}
	if (channelChatID.ID == 0 && channelChatID.Username == "") || (artwork.Pictures[index].TelegramInfo == nil || artwork.Pictures[index].TelegramInfo.MessageID == 0) {
		return []telego.InlineKeyboardButton{
			telegoutil.InlineKeyboardButton("详情").WithURL(config.Cfg.API.SiteURL + "/artwork/" + artwork.ID),
			telegoutil.InlineKeyboardButton("原图").WithURL(GetDeepLinkForFile(artwork.Pictures[index].ID, botUsername)),
		}
	}
	return []telego.InlineKeyboardButton{
		telegoutil.InlineKeyboardButton("详情").WithURL(GetArtworkPostMessageURL(artwork.Pictures[index].TelegramInfo.MessageID, channelChatID)),
		telegoutil.InlineKeyboardButton("原图").WithURL(GetDeepLinkForFile(artwork.Pictures[index].ID, botUsername)),
	}
}

func GetMessagePhotoFileBytes(bot *telego.Bot, message *telego.Message) ([]byte, error) {
	fileID := ""
	if message.Photo != nil {
		fileID = message.Photo[len(message.Photo)-1].FileID
	}
	if message.Document != nil && strings.HasPrefix(message.Document.MimeType, "image/") {
		if message.Document.FileSize > 20*1024*1024 {
			return nil, errors.ErrFileTooLarge
		}
		fileID = message.Document.FileID
	}
	if fileID == "" {
		return nil, errors.ErrNoPhotoInMessage
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
			return nil, err
		}
		file = telegoutil.File(telegoutil.NameReader(bytes.NewReader(data), filepath.Base(picture.StorageInfo.Original.Path)))
	}
	document := telegoutil.Document(message.Chat.ChatID(), file).
		WithReplyParameters(&telego.ReplyParameters{
			MessageID: message.MessageID,
		})
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
				Logger.Warnf("获取作品信息失败: %s", err)
			}
		} else {
			Logger.Warnf("创建 ObjectID 失败: %s", err)
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
				Logger.Warnf("更新图片信息失败: %s", err)
			}
		}
	}
	return documentMessage, nil
}

func GetPicturePreviewInputFile(ctx context.Context, picture *types.Picture) (inputFile *telego.InputFile, needUpdatePreview bool, err error) {
	if picture.TelegramInfo != nil && picture.TelegramInfo.PhotoFileID != "" {
		inputFileStruct := telegoutil.FileFromID(picture.TelegramInfo.PhotoFileID)
		inputFile = &inputFileStruct
		return
	}
	if fileBytes := common.GetReqCachedFile(picture.Original); fileBytes != nil {
		fileBytes, err = common.CompressImageToJPEG(fileBytes, 10, 2560, picture.Original)
		if err != nil {
			return nil, false, err
		}
		inputFileStruct := telegoutil.File(telegoutil.NameReader(bytes.NewReader(fileBytes), picture.Original))
		inputFile = &inputFileStruct
		return
	}
	needUpdatePreview = true
	if botPhotoFileID := service.GetEtcData(ctx, "bot_photo_file_id"); botPhotoFileID != nil {
		inputFileStruct := telegoutil.FileFromID(botPhotoFileID.(string))
		inputFile = &inputFileStruct
		return
	}
	if botPhotoFileBytes := service.GetEtcData(ctx, "bot_photo_bytes"); botPhotoFileBytes != nil {
		if data, ok := botPhotoFileBytes.(primitive.Binary); ok {
			inputFileStruct := telegoutil.File(telegoutil.NameReader(bytes.NewReader(data.Data), picture.Original))
			inputFile = &inputFileStruct
			return
		}
	}
	return nil, true, errors.ErrNoAvailableFile
}
