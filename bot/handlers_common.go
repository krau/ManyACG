package bot

import (
	"ManyACG-Bot/service"
	"ManyACG-Bot/sources"
	"ManyACG-Bot/telegram"
	"ManyACG-Bot/types"
	"context"
	"errors"
	"math/rand"
	"strconv"
	"strings"

	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegoutil"
	"go.mongodb.org/mongo-driver/mongo"

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
				telegram.ReplyMessage(bot, message, "获取失败: "+err.Error())
				return
			}
			_, err = telegram.SendPictureFileByMessageID(ctx, bot, message, messageID)
			if err != nil {
				telegram.ReplyMessage(bot, message, "获取失败: "+err.Error())
				return
			}
		}
		return
	}
	help(ctx, bot, message)
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
	helpText += "源码: https://github.com/krau/ManyACG-Bot"
	telegram.ReplyMessage(bot, message, helpText)
}

func getPictureFile(ctx context.Context, bot *telego.Bot, message telego.Message) {
	replyToMessage := message.ReplyToMessage
	if replyToMessage == nil {
		telegram.ReplyMessage(bot, message, "请使用该命令回复一条频道的消息")
		return
	}
	if replyToMessage.Photo == nil {
		telegram.ReplyMessage(bot, message, "目标消息不包含图片")
		return
	}
	if replyToMessage.ForwardOrigin == nil {
		telegram.ReplyMessage(bot, message, "请使用该命令回复一条频道的消息")
		return
	}

	var messageOriginChannel *telego.MessageOriginChannel
	if replyToMessage.ForwardOrigin.OriginType() == telego.OriginTypeChannel {
		messageOriginChannel = replyToMessage.ForwardOrigin.(*telego.MessageOriginChannel)
	} else {
		telegram.ReplyMessage(bot, message, "请使用该命令回复一条频道的消息")
		return
	}

	_, err := telegram.SendPictureFileByMessageID(ctx, bot, message, messageOriginChannel.MessageID)
	if err != nil {
		telegram.ReplyMessage(bot, message, "获取失败: "+err.Error())
		return
	}
}

func randomPicture(ctx context.Context, bot *telego.Bot, message telego.Message) {
	cmd, _, args := telegoutil.ParseCommand(message.Text)
	r18 := cmd == "setu"
	limit := 1
	Logger.Infof("randomPicture: r18=%v, args=%v", r18, args)
	artwork, err := service.GetRandomArtworksByTagsR18(ctx, args, r18, limit)
	if err != nil {
		Logger.Warnf("获取图片失败: %s", err)
		text := "获取图片失败" + err.Error()
		if errors.Is(err, mongo.ErrNoDocuments) {
			text = "未找到图片"
		}
		telegram.ReplyMessage(bot, message, text)
		return
	}
	if len(artwork) == 0 {
		telegram.ReplyMessage(bot, message, "未找到图片")
		return
	}
	pictures := artwork[0].Pictures
	picture := pictures[rand.Intn(len(pictures))]
	var file telego.InputFile
	if picture.TelegramInfo.PhotoFileID != "" {
		file = telegoutil.FileFromID(picture.TelegramInfo.PhotoFileID)
	} else {
		photoURL := picture.Original
		if artwork[0].SourceType == types.SourceTypePixiv {
			photoURL = sources.GetPixivRegularURL(photoURL)
		}
		file = telegoutil.FileFromURL(photoURL)
	}
	photoMessage, err := bot.SendPhoto(telegoutil.Photo(message.Chat.ChatID(), file).
		WithReplyParameters(&telego.ReplyParameters{
			MessageID: message.MessageID,
		}).WithCaption(artwork[0].Title).WithReplyMarkup(
		telegoutil.InlineKeyboard([]telego.InlineKeyboardButton{
			telegoutil.InlineKeyboardButton("来源").WithURL(telegram.GetArtworkPostMessageURL(picture.TelegramInfo.MessageID)),
			telegoutil.InlineKeyboardButton("原图").WithURL(telegram.GetDeepLinkForFile(picture.TelegramInfo.MessageID)),
		}),
	))
	if err != nil {
		telegram.ReplyMessage(bot, message, "发送图片失败: "+err.Error())
	}
	if photoMessage != nil {
		picture.TelegramInfo.PhotoFileID = photoMessage.Photo[len(photoMessage.Photo)-1].FileID
		if service.UpdatePictureTelegramInfo(ctx, picture, picture.TelegramInfo) != nil {
			Logger.Warnf("更新图片信息失败: %s", err)
		}
	}
}

func getArtworkInfo(ctx context.Context, bot *telego.Bot, message telego.Message) {
	if service.CheckAdminPermission(ctx, message.From.ID, types.GetArtworkInfo) {
		getArtworkInfoForAdmin(ctx, bot, message)
		return
	}
	sourceURL := ""
	if message.Caption != "" {
		sourceURL = sources.MatchSourceURL(message.Caption)
	} else {
		sourceURL = sources.MatchSourceURL(message.Text)
	}
	if sourceURL == "" {
		return
	}
	artwork, err := service.GetArtworkByURL(ctx, sourceURL)
	if err != nil {
		Logger.Warnf("获取图片信息失败: %s", err)
		return
	}
	var inputFile telego.InputFile
	if artwork.Pictures[0].TelegramInfo == nil || artwork.Pictures[0].TelegramInfo.PhotoFileID == "" {
		photoURL := artwork.Pictures[0].Original
		if artwork.SourceType == types.SourceTypePixiv {
			photoURL = sources.GetPixivRegularURL(photoURL)
		}
		inputFile = telegoutil.FileFromURL(photoURL)
	} else {
		inputFile = telegoutil.FileFromID(artwork.Pictures[0].TelegramInfo.PhotoFileID)
	}
	photo := telegoutil.Photo(message.Chat.ChatID(), inputFile).
		WithReplyParameters(&telego.ReplyParameters{MessageID: message.MessageID}).
		WithParseMode(telego.ModeMarkdownV2).
		WithCaption(telegram.GetArtworkMarkdownCaption(artwork)).
		WithReplyMarkup(telegoutil.InlineKeyboard(
			[]telego.InlineKeyboardButton{
				telegoutil.InlineKeyboardButton("来源").WithURL(telegram.GetArtworkPostMessageURL(artwork.Pictures[0].TelegramInfo.MessageID)),
				telegoutil.InlineKeyboardButton("原图").WithURL(telegram.GetDeepLinkForFile(artwork.Pictures[0].TelegramInfo.MessageID)),
			},
		))

	if artwork.R18 {
		photo.WithHasSpoiler()
	}
	bot.SendPhoto(photo)
}
