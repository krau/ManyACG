package telegram

import (
	"ManyACG-Bot/service"
	"ManyACG-Bot/storage"
	"bytes"
	"context"
	"path/filepath"
	"regexp"
	"strings"

	. "ManyACG-Bot/logger"

	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegoutil"
)

func escapeMarkdown(text string) string {
	escapeChars := `\_*[]()~` + "`" + ">#+-=|{}.!"
	re := regexp.MustCompile("([" + regexp.QuoteMeta(escapeChars) + "])")
	return re.ReplaceAllString(text, "\\$1")
}

func replaceChars(input string, oldChars []string, newChar string) string {
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

func sendPictureFileByMessageID(ctx context.Context, bot *telego.Bot, message telego.Message, pictureMessageID int) (*telego.Message, error) {
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
		}))
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
