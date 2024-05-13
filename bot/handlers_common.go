package bot

import (
	"ManyACG/common"
	"ManyACG/service"
	"ManyACG/sources"
	"ManyACG/telegram"
	"ManyACG/types"
	"context"
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	. "ManyACG/logger"

	"github.com/google/uuid"
	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegoutil"
	"go.mongodb.org/mongo-driver/mongo"
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
/search - 搜索相似图片
`
	isAdmin, _ := service.IsAdmin(ctx, message.From.ID)
	if isAdmin {
		helpText += `/set_admin - 设置|删除管理员
/del - 删除图片
/fetch - 手动开始一次抓取
/process_pictures - 处理无哈希的图片

发送作品链接可以获取信息或发布到频道
`
	}
	helpText += "源码: https://github.com/krau/ManyACG"
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
	photo := telegoutil.Photo(message.Chat.ChatID(), file).
		WithReplyParameters(&telego.ReplyParameters{
			MessageID: message.MessageID,
		}).WithCaption(artwork[0].Title).WithReplyMarkup(
		telegoutil.InlineKeyboard([]telego.InlineKeyboardButton{
			telegoutil.InlineKeyboardButton("来源").WithURL(telegram.GetArtworkPostMessageURL(picture.TelegramInfo.MessageID)),
			telegoutil.InlineKeyboardButton("原图").WithURL(telegram.GetDeepLinkForFile(picture.TelegramInfo.MessageID)),
		}),
	)
	if artwork[0].R18 {
		photo.WithHasSpoiler()
	}
	photoMessage, err := bot.SendPhoto(photo)
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
	hasPermission := CheckPermissionInGroup(ctx, message, types.GetArtworkInfo)
	sourceURL := FindSourceURLForMessage(&message)
	var waitMessageID int
	if hasPermission {
		go func() {
			msg, err := telegram.ReplyMessage(bot, message, "正在获取作品信息...")
			if err != nil {
				Logger.Warnf("发送消息失败: %s", err)
				return
			}
			waitMessageID = msg.MessageID
		}()
	}

	defer func() {
		if r := recover(); r != nil {
			Logger.Fatalf("panic: %v", r)
		}
		time.Sleep(1 * time.Second)
		if waitMessageID != 0 {
			go bot.DeleteMessage(telegoutil.Delete(message.Chat.ChatID(), waitMessageID))
		}
	}()

	isAlreadyPosted := true
	artwork, err := service.GetArtworkByURL(ctx, sourceURL)
	if err != nil {
		if !hasPermission {
			return
		}
		isAlreadyPosted = false
		artwork, err = sources.GetArtworkInfo(sourceURL)
		if err != nil {
			telegram.ReplyMessage(bot, message, "获取作品信息失败: "+err.Error())
			return
		}
		if err := service.CreateCachedArtwork(ctx, artwork); err != nil {
			Logger.Warnf("缓存作品失败: %s", err)
		}
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
	}
}

func searchPicture(ctx context.Context, bot *telego.Bot, message telego.Message) {
	hasPermission := CheckPermissionInGroup(ctx, message, types.SearchPicture)
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
	if len(pictures) > 0 {
		text := fmt.Sprintf("找到%d张相似或相同的图片\n\n", len(pictures))
		for _, picture := range pictures {
			artwork, err := service.GetArtworkByMessageID(ctx, picture.TelegramInfo.MessageID)
			if err != nil {
				text += telegram.EscapeMarkdown(fmt.Sprintf("%s 模糊度: %.2f\n\n", telegram.GetArtworkPostMessageURL(picture.TelegramInfo.MessageID), picture.BlurScore))
			}
			text += fmt.Sprintf("[%s\\_%d](%s)  ",
				telegram.EscapeMarkdown(artwork.Title),
				picture.Index+1,
				telegram.EscapeMarkdown(telegram.GetArtworkPostMessageURL(picture.TelegramInfo.MessageID)))
			text += telegram.EscapeMarkdown(fmt.Sprintf("模糊度: %.2f\n\n", picture.BlurScore))
		}
		telegram.ReplyMessageWithMarkdown(bot, message, text)
		return
	}
	if !hasPermission {
		telegram.ReplyMessage(bot, message, "未在数据库中找到相似图片")
		return
	}
	// TODO: 有权限时使用其他搜索引擎搜图
	telegram.ReplyMessage(bot, message, "未找到相似图片")
}

func inlineQuery(ctx context.Context, bot *telego.Bot, query telego.InlineQuery) {
	queryText := query.Query
	tags := strings.Split(queryText, " ")
	r18 := false
	for i, tag := range tags {
		if tag == "r18" {
			r18 = true
			tags = append(tags[:i], tags[i+1:]...)
			break
		}
	}
	limit := 44
	artworks, err := service.GetRandomArtworksByTagsR18(ctx, tags, r18, limit)
	if err != nil {
		bot.AnswerInlineQuery(telegoutil.InlineQuery(query.ID, telegoutil.ResultArticle(uuid.NewString(), "未找到相关图片", telegoutil.TextMessage("/setu"))))
		return
	}
	results := make([]telego.InlineQueryResult, 0, len(artworks))
	for _, artwork := range artworks {
		picture := artwork.Pictures[rand.Intn(len(artwork.Pictures))]
		if picture.TelegramInfo.PhotoFileID == "" {
			continue
		}
		result := telegoutil.ResultCachedPhoto(uuid.NewString(), picture.TelegramInfo.PhotoFileID).WithCaption(artwork.Title)
		result.WithReplyMarkup(telegoutil.InlineKeyboard([]telego.InlineKeyboardButton{
			telegoutil.InlineKeyboardButton("来源").WithURL(telegram.GetArtworkPostMessageURL(picture.TelegramInfo.MessageID)),
			telegoutil.InlineKeyboardButton("原图").WithURL(telegram.GetDeepLinkForFile(picture.TelegramInfo.MessageID)),
		}))
		results = append(results, result)
	}
	bot.AnswerInlineQuery(telegoutil.InlineQuery(query.ID, results...).WithCacheTime(3))
}
