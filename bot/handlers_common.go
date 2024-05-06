package bot

import (
	"ManyACG-Bot/common"
	"ManyACG-Bot/service"
	"ManyACG-Bot/sources"
	"ManyACG-Bot/telegram"
	"ManyACG-Bot/types"
	"context"
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

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
	if !CheckPermissionInGroup(ctx, message, types.GetArtworkInfo) {
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

	var waitMessageID int
	go func() {
		msg, err := telegram.ReplyMessage(bot, message, "正在获取作品信息...")
		if err != nil {
			Logger.Warnf("发送消息失败: %s", err)
			return
		}
		waitMessageID = msg.MessageID
	}()

	defer func() {
		if r := recover(); r != nil {
			Logger.Errorf("panic: %v", r)
		}
		time.Sleep(1 * time.Second)
		if waitMessageID != 0 {
			go bot.DeleteMessage(telegoutil.Delete(message.Chat.ChatID(), waitMessageID))
		}
	}()

	isAlreadyPosted := true
	artwork, err := service.GetArtworkByURL(ctx, sourceURL)
	if err != nil {
		isAlreadyPosted = false
		artwork, err = sources.GetArtworkInfo(sourceURL)
	}
	if err != nil {
		telegram.ReplyMessage(bot, message, "获取作品信息失败: "+err.Error())
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
		WithParseMode(telego.ModeMarkdownV2)

	deletedModel, _ := service.GetDeletedByURL(ctx, sourceURL)
	artworkInfoCaption := telegram.GetArtworkMarkdownCaption(artwork)
	if deletedModel != nil {
		photo.WithCaption(artworkInfoCaption + telegram.EscapeMarkdown(fmt.Sprintf("\n\n这是一个在 %s 删除的作品\n\n"+
			"如果发布则会取消删除", deletedModel.DeletedAt.Time().Format("2006-01-02 15:04:05"))))
	} else {
		photo.WithCaption(telegram.GetArtworkMarkdownCaption(artwork))
	}
	if isAlreadyPosted {
		photo.WithReplyMarkup(telegoutil.InlineKeyboard(
			[]telego.InlineKeyboardButton{
				telegoutil.InlineKeyboardButton("来源").WithURL(telegram.GetArtworkPostMessageURL(artwork.Pictures[0].TelegramInfo.MessageID)),
				telegoutil.InlineKeyboardButton("原图").WithURL(telegram.GetDeepLinkForFile(artwork.Pictures[0].TelegramInfo.MessageID)),
			},
		))
	} else {
		id, err := service.CreateCallbackData(ctx, artwork.SourceURL)
		if err != nil {
			telegram.ReplyMessage(bot, message, "创建回调数据失败: "+err.Error())
			return
		}
		photo.WithReplyMarkup(telegoutil.InlineKeyboard(
			[]telego.InlineKeyboardButton{
				telegoutil.InlineKeyboardButton("发布到频道").WithCallbackData("admin post_artwork " + id),
			},
		))
	}
	if artwork.R18 {
		photo.WithHasSpoiler()
	}
	_, err = bot.SendPhoto(photo)
	if err != nil {
		telegram.ReplyMessage(bot, message, "发送图片失败: "+err.Error())
		return
	}
}

func searchPicture(ctx context.Context, bot *telego.Bot, message telego.Message) {
	if !CheckPermissionInGroup(ctx, message, types.SearchPicture) {
		telegram.ReplyMessage(bot, message, "用户或群组没有权限")
		return
	}
	if message.ReplyToMessage == nil {
		telegram.ReplyMessage(bot, message, "请使用该命令回复一条图片消息")
		return
	}
	if message.ReplyToMessage.Photo == nil {
		telegram.ReplyMessage(bot, message, "目标消息不包含图片")
		return
	}
	go telegram.ReplyMessage(bot, message, "少女祈祷中...")
	photo := message.ReplyToMessage.Photo
	photoFileID := photo[len(photo)-1].FileID
	tgFile, err := bot.GetFile(&telego.GetFileParams{
		FileID: photoFileID,
	})
	if err != nil {
		telegram.ReplyMessage(bot, message, "获取文件信息失败: "+err.Error())
		return
	}
	fileBytes, err := telegoutil.DownloadFile(bot.FileDownloadURL(tgFile.FilePath))
	if err != nil {
		telegram.ReplyMessage(bot, message, "下载文件失败: "+err.Error())
		return
	}
	hash, err := common.GetPhash(fileBytes)
	if err != nil {
		telegram.ReplyMessage(bot, message, "获取图片哈希失败: "+err.Error())
		return
	}
	pictures, err := service.GetPicturesByHash(ctx, hash)
	if err != nil {
		telegram.ReplyMessage(bot, message, "搜索图片失败: "+err.Error())
		return
	}
	if len(pictures) == 0 {
		telegram.ReplyMessage(bot, message, "未在数据库中找到图片")
		return
	}
	text := fmt.Sprintf("找到%d张相似或相同图片\n", len(pictures))
	for i, picture := range pictures {
		artwork, err := service.GetArtworkByMessageID(ctx, picture.TelegramInfo.MessageID)
		if err != nil {
			text += fmt.Sprintf("%d\. %s\n", i+1, telegram.EscapeMarkdown(telegram.GetArtworkPostMessageURL(picture.TelegramInfo.MessageID)))
			continue
		}
		text += fmt.Sprintf("%d\. [%s](%s)\n", i+1, artwork.Title, telegram.EscapeMarkdown(telegram.GetArtworkPostMessageURL(picture.TelegramInfo.MessageID)))
	}
	bot.SendMessage(telegoutil.Message(message.Chat.ChatID(), text).
		WithReplyParameters(&telego.ReplyParameters{
			MessageID: message.MessageID,
		}).WithParseMode(telego.ModeMarkdownV2))
}
