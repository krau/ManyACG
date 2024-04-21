package bot

import (
	"ManyACG-Bot/common"
	"ManyACG-Bot/service"
	"ManyACG-Bot/telegram"
	"ManyACG-Bot/types"
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegoutil"

	. "ManyACG-Bot/logger"
)

func start(ctx context.Context, bot *telego.Bot, message telego.Message) {
	_, _, args := telegoutil.ParseCommand(message.Text)
	if len(args) > 0 {
		Logger.Debugf("start: args=%v", args)
		if strings.HasPrefix(args[0], "file_") {
			messageIDStr := args[0][5:]
			messageID, err := strconv.Atoi(messageIDStr)
			if err != nil {
				bot.SendMessage(telegoutil.Messagef(message.Chat.ChatID(), "获取失败: %s", err).WithReplyParameters(
					&telego.ReplyParameters{
						MessageID: message.MessageID,
					},
				))
				return
			}
			_, err = telegram.SendPictureFileByMessageID(ctx, bot, message, messageID)
			if err != nil {
				bot.SendMessage(telegoutil.Messagef(message.Chat.ChatID(), "获取失败: %s", err).WithReplyParameters(
					&telego.ReplyParameters{
						MessageID: message.MessageID,
					},
				))
				return
			}
		}
		return
	}

	bot.SendMessage(
		telegoutil.Message(message.Chat.ChatID(),
			"喵喵喵~\n这是 ManyACG-Bot 的一个实例\n\n源码: https://github.com/krau/ManyACG-Bot").
			WithReplyParameters(&telego.ReplyParameters{
				MessageID: message.MessageID,
			}))
}

func help(ctx context.Context, bot *telego.Bot, message telego.Message) {
	helpText := `使用方法:
/start - 喵喵喵
/file - 回复一条频道的消息获取原图文件
/setu - 来点涩图 <tag1> <tag2> ...
/random - 随机1张全年龄图片 <tag1> <tag2> ...
`
	isAdmin, _ := service.IsAdmin(ctx, message.From.ID)
	if isAdmin {
		helpText += `/set_admin - 设置|删除管理员
/del - 删除图片
/fetch - 手动开始一次抓取

发送作品链接可以获取信息或发布到频道
`
	}
	bot.SendMessage(telegoutil.Message(message.Chat.ChatID(), helpText).
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

	_, err := telegram.SendPictureFileByMessageID(ctx, bot, message, messageOriginChannel.MessageID)
	if err != nil {
		bot.SendMessage(telegoutil.Messagef(message.Chat.ChatID(), "获取失败: %s", err).WithReplyParameters(
			&telego.ReplyParameters{
				MessageID: message.MessageID,
			},
		))
		return
	}
}

func randomPicture(ctx context.Context, bot *telego.Bot, message telego.Message) {
	cmd, _, args := telegoutil.ParseCommand(message.Text)
	r18 := cmd == "setu"
	limit := 1
	Logger.Debugf("randomPicture: r18=%v, args=%v", r18, args)
	artwork, err := service.GetRandomArtworksByTagsR18(ctx, args, r18, limit)
	if err != nil {
		Logger.Warnf("获取图片失败: %s", err)
		bot.SendMessage(telegoutil.Messagef(message.Chat.ChatID(), "获取图片失败: %s", err).
			WithReplyParameters(&telego.ReplyParameters{
				MessageID: message.MessageID,
			}))
		return
	}
	if len(artwork) == 0 {
		bot.SendMessage(telegoutil.Message(message.Chat.ChatID(), "未找到图片").
			WithReplyParameters(&telego.ReplyParameters{
				MessageID: message.MessageID,
			}))
		return
	}
	picture := artwork[0].Pictures[0]
	var file telego.InputFile
	if picture.TelegramInfo.PhotoFileID != "" {
		file = telegoutil.FileFromID(picture.TelegramInfo.PhotoFileID)
	} else {
		photoURL := picture.Original
		if artwork[0].SourceType == types.SourceTypePixiv {
			photoURL = common.GetPixivRegularURL(photoURL)
		}
		file = telegoutil.FileFromURL(photoURL)
	}
	photoMessage, err := bot.SendPhoto(telegoutil.Photo(message.Chat.ChatID(), file).
		WithReplyParameters(&telego.ReplyParameters{
			MessageID: message.MessageID,
		}).WithCaption(artwork[0].Title).WithReplyMarkup(
		telegoutil.InlineKeyboard([]telego.InlineKeyboardButton{
			telegoutil.InlineKeyboardButton("来源").WithURL(fmt.Sprintf("https://t.me/%s/%d", strings.ReplaceAll(telegram.ChannelChatID.String(), "@", ""), picture.TelegramInfo.MessageID)),
			telegoutil.InlineKeyboardButton("原图").WithURL(fmt.Sprintf("https://t.me/%s/?start=file_%d", telegram.BotUsername, picture.TelegramInfo.MessageID)),
		}),
	))
	if err != nil {
		bot.SendMessage(telegoutil.Messagef(message.Chat.ChatID(), "发送图片失败: %s", err).
			WithReplyParameters(&telego.ReplyParameters{
				MessageID: message.MessageID,
			}))
	}
	if photoMessage != nil {
		picture.TelegramInfo.PhotoFileID = photoMessage.Photo[len(photoMessage.Photo)-1].FileID
		if service.UpdatePictureTelegramInfo(ctx, picture, picture.TelegramInfo) != nil {
			Logger.Warnf("更新图片信息失败: %s", err)
		}
	}
}
