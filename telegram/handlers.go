package telegram

import (
	"ManyACG-Bot/service"
	"ManyACG-Bot/storage"
	"bytes"
	"context"
	"path/filepath"

	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegoutil"

	. "ManyACG-Bot/logger"
)

func start(_ context.Context, bot *telego.Bot, message telego.Message) {
	bot.SendMessage(
		telegoutil.Message(message.Chat.ChatID(),
			"Hi~").
			WithReplyParameters(&telego.ReplyParameters{
				MessageID: message.MessageID,
			}))
}

func getPictureFile(ctx context.Context, bot *telego.Bot, message telego.Message) {
	replyToMessage := message.ReplyToMessage
	if replyToMessage == nil {
		bot.SendMessage(telegoutil.Message(message.Chat.ChatID(), "请使用该命令回复一条频道的消息").
			WithReplyParameters(&telego.ReplyParameters{
				MessageID: message.MessageID,
			}))
		return
	}
	if replyToMessage.Photo == nil {
		bot.SendMessage(telegoutil.Message(message.Chat.ChatID(), "目标消息不包含图片").
			WithReplyParameters(&telego.ReplyParameters{
				MessageID: message.MessageID,
			}))
		return
	}
	if replyToMessage.ForwardOrigin == nil {
		bot.SendMessage(telegoutil.Message(message.Chat.ChatID(), "请使用该命令回复一条频道的消息").
			WithReplyParameters(&telego.ReplyParameters{
				MessageID: message.MessageID,
			}))
		return
	}

	var messageOriginChannel *telego.MessageOriginChannel
	if replyToMessage.ForwardOrigin.OriginType() == telego.OriginTypeChannel {
		messageOriginChannel = replyToMessage.ForwardOrigin.(*telego.MessageOriginChannel)
	} else {
		bot.SendMessage(telegoutil.Message(message.Chat.ChatID(), "请使用该命令回复一条频道的消息").
			WithReplyParameters(&telego.ReplyParameters{
				MessageID: message.MessageID,
			}))
		return
	}

	picture, err := service.GetPictureByMessageID(ctx, messageOriginChannel.MessageID)
	if err != nil {
		bot.SendMessage(telegoutil.Messagef(message.Chat.ChatID(), "查询图片信息失败: %s", err).
			WithReplyParameters(&telego.ReplyParameters{
				MessageID: message.MessageID,
			}))
		return
	}

	var file telego.InputFile
	if picture.TelegramInfo.DocumentFileID != "" {
		file = telegoutil.FileFromID(picture.TelegramInfo.DocumentFileID)
	} else {
		data, err := storage.GetStorage().GetFile(picture.StorageInfo)
		if err != nil {
			bot.SendMessage(telegoutil.Messagef(message.Chat.ChatID(), "获取图片失败: %s", err).
				WithReplyParameters(&telego.ReplyParameters{
					MessageID: message.MessageID,
				}))
			return
		}
		file = telegoutil.File(telegoutil.NameReader(bytes.NewReader(data), filepath.Base(picture.Original)))
	}
	documentMessage, err := bot.SendDocument(telegoutil.Document(message.Chat.ChatID(), file).
		WithReplyParameters(&telego.ReplyParameters{
			MessageID: message.MessageID,
		}))
	if err != nil {
		bot.SendMessage(telegoutil.Messagef(ChatID, "发送文件失败: %s", err))
	}
	if documentMessage != nil {
		if documentMessage.Document != nil {
			picture.TelegramInfo.DocumentFileID = documentMessage.Document.FileID
			if service.UpdatePictureTelegramInfo(ctx, picture, picture.TelegramInfo) != nil {
				Logger.Warnf("更新图片信息失败: %s", err)
			}
		}
	}

}
