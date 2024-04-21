package telegram

import (
	"ManyACG-Bot/service"
	"ManyACG-Bot/storage"
	"ManyACG-Bot/types"
	"bytes"
	"context"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	. "ManyACG-Bot/logger"

	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegoutil"
)

func EscapeMarkdown(text string) string {
	escapeChars := `\_*[]()~` + "`" + ">#+-=|{}.!"
	re := regexp.MustCompile("([" + regexp.QuoteMeta(escapeChars) + "])")
	return re.ReplaceAllString(text, "\\$1")
}

func ReplaceChars(input string, oldChars []string, newChar string) string {
	for _, char := range oldChars {
		input = strings.ReplaceAll(input, char, newChar)
	}
	return input
}

func GetMessageIDs(messages []telego.Message) []int {
	ids := make([]int, len(messages))
	for i, message := range messages {
		ids[i] = message.MessageID
	}
	return ids
}

func SendPictureFileByMessageID(ctx context.Context, bot *telego.Bot, message telego.Message, pictureMessageID int) (*telego.Message, error) {
	picture, err := service.GetPictureByMessageID(ctx, pictureMessageID)
	if err != nil {
		return nil, err
	}
	var file telego.InputFile
	if picture.TelegramInfo.DocumentFileID != "" {
		file = telegoutil.FileFromID(picture.TelegramInfo.DocumentFileID)
	} else {
		data, err := storage.GetStorage().GetFile(picture.StorageInfo)
		if err != nil {
			return nil, err
		}
		file = telegoutil.File(telegoutil.NameReader(bytes.NewReader(data), filepath.Base(picture.Original)))
	}

	documentMessage, err := bot.SendDocument(telegoutil.Document(message.Chat.ChatID(), file).
		WithReplyParameters(&telego.ReplyParameters{
			MessageID: message.MessageID,
		}).WithCaption("这是你要的原图~"))
	if err != nil {
		return nil, err
	}
	if documentMessage != nil {
		if documentMessage.Document != nil {
			picture.TelegramInfo.DocumentFileID = documentMessage.Document.FileID
			if service.UpdatePictureTelegramInfo(ctx, picture, picture.TelegramInfo) != nil {
				Logger.Warnf("更新图片信息失败: %s", err)
			}
		}
	}
	return documentMessage, nil
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

func CheckTargetMessageIsChannelArtworkPost(ctx context.Context, bot *telego.Bot, message telego.Message) (*telego.MessageOriginChannel, bool) {
	if message.ReplyToMessage == nil {
		return nil, false
	}
	if !message.ReplyToMessage.IsAutomaticForward || message.ReplyToMessage.Photo == nil || message.ReplyToMessage.ForwardOrigin == nil {
		return nil, false
	}
	messageOriginChannel := GetMssageOriginChannel(ctx, bot, *message.ReplyToMessage)
	if messageOriginChannel == nil {
		return nil, false
	}
	return messageOriginChannel, true
}

func GetArtworkMarkdownCaption(artwork *types.Artwork) string {
	caption := fmt.Sprintf("[*%s*](%s)", (artwork.Title), artwork.SourceURL)
	caption += "\n\n" + "*Author:* " + EscapeMarkdown(artwork.Artist.Name)
	if artwork.Description != "" {
		if len(artwork.Description) > 233 {
			caption += "\n\n" + "_" + EscapeMarkdown(artwork.Description[:233]) + "\\.\\.\\._"
		} else {
			caption += "\n\n" + "_" + EscapeMarkdown(artwork.Description) + "_"
		}
	}
	tags := ""
	for _, tag := range artwork.Tags {
		tag = ReplaceChars(tag, []string{":", "：", "-", "（", "）", "「", "」", "*"}, "_")
		tag = ReplaceChars(tag, []string{"?"}, "")
		tag = ReplaceChars(tag, []string{"/"}, " "+"#")
		tags += "\\#" + strings.Join(strings.Split(EscapeMarkdown(tag), " "), "") + " "
	}
	caption += "\n\n" + tags
	return caption
}
