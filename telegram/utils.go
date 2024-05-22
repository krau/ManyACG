package telegram

import (
	"ManyACG/common"
	"ManyACG/service"
	"ManyACG/sources"
	"ManyACG/storage"
	"ManyACG/types"
	"bytes"
	"context"
	"fmt"
	"strings"

	. "ManyACG/logger"

	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegoutil"
)

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
		go ReplyMessage(bot, message, "正在下载原图，请稍等~")
		data, err := storage.GetStorage().GetFile(picture.StorageInfo)
		if err != nil {
			return nil, err
		}
		artwork, err := service.GetArtworkByMessageID(ctx, pictureMessageID)
		if err != nil {
			return nil, err
		}
		file = telegoutil.File(telegoutil.NameReader(bytes.NewReader(data), sources.GetFileName(artwork, picture)))
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

func GetArtworkMarkdownCaption(artwork *types.Artwork) string {
	caption := fmt.Sprintf("[*%s*](%s)", common.EscapeMarkdown(artwork.Title), artwork.SourceURL)
	caption += "\n" + "*Author:* " + common.EscapeMarkdown(artwork.Artist.Name)
	if artwork.Description != "" {
		desc := artwork.Description
		lines := strings.Split(desc, "\n")
		for i, line := range lines {
			lines[i] = ">" + common.EscapeMarkdown(line)
			if i == len(lines)-1 {
				lines[i] += "**"
			}
		}
		desc = strings.Join(lines, "\n")
		if len(desc) > 500 {
			caption += "\n\n" + desc[:500] + "\\.\\.\\."
		} else {
			caption += "\n\n" + desc
		}
	}
	tags := ""
	for _, tag := range artwork.Tags {
		tag = ReplaceChars(tag, []string{":", "：", "-", "（", "）", "「", "」", "*"}, "_")
		tag = ReplaceChars(tag, []string{"?"}, "")
		tag = ReplaceChars(tag, []string{"/"}, " "+"#")
		tags += "\\#" + strings.Join(strings.Split(common.EscapeMarkdown(tag), " "), "") + " "
	}
	caption += "\n\n" + tags
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

func GetArtworkPostMessageURL(messageID int) string {
	return fmt.Sprintf("https://t.me/%s/%d", strings.ReplaceAll(ChannelChatID.String(), "@", ""), messageID)
}

func GetDeepLinkForFile(messageID int) string {
	return fmt.Sprintf("https://t.me/%s/?start=file_%d", BotUsername, messageID)
}

func GetPostedPictureReplyMarkup(picture *types.Picture) telego.ReplyMarkup {
	return telegoutil.InlineKeyboard(
		GetPostedPictureInlineKeyboardButton(picture),
	)
}

func GetPostedPictureInlineKeyboardButton(picture *types.Picture) []telego.InlineKeyboardButton {
	return []telego.InlineKeyboardButton{
		telegoutil.InlineKeyboardButton("详情").WithURL(GetArtworkPostMessageURL(picture.TelegramInfo.MessageID)),
		telegoutil.InlineKeyboardButton("原图").WithURL(GetDeepLinkForFile(picture.TelegramInfo.MessageID)),
	}
}
