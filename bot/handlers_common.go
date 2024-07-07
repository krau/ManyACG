package bot

import (
	"ManyACG/common"
	"ManyACG/config"
	"ManyACG/service"
	"ManyACG/sources"
	"ManyACG/storage"
	"ManyACG/telegram"
	"ManyACG/types"
	"bytes"
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
/file - 回复一条频道的消息获取原图文件 <index>
/setu - 随机图片(NSFW)
/random - 随机全年龄图片
/search - 搜索相似图片

关键词参数使用 '|' 分隔或关系, 使用空格分隔与关系, 示例:

/random 萝莉|白丝 猫耳|原创

表示搜索包含"萝莉"或"白丝", 且包含"猫耳"或"原创"的图片.
Inline 查询支持同样的参数格式.

`
	isAdmin, _ := service.IsAdmin(ctx, message.From.ID)
	if isAdmin {
		helpText += `/set_admin - 设置|删除管理员
/del - 删除图片 <消息id>
/delete - 删除整个作品
/r18 - 设置作品R18标记
/tags - 更新作品标签(覆盖原有标签)
/addtags - 添加作品标签
/deltags - 删除作品标签
/fetch - 手动开始一次抓取
/process_pictures - 处理无哈希的图片

发送作品链接可以获取信息或发布到频道
`
	}
	helpText += "源码: https://github.com/krau/ManyACG"
	telegram.ReplyMessage(bot, message, helpText)
}

func getPictureFile(ctx context.Context, bot *telego.Bot, message telego.Message) {
	messageOrigin, ok := telegram.GetMessageOriginChannelArtworkPost(ctx, bot, message)
	if !ok {
		if message.ReplyToMessage == nil {
			telegram.ReplyMessage(bot, message, "请回复一条频道的图片消息")
			return
		}
		fileBytes, err := telegram.GetMessagePhotoFileBytes(bot, message.ReplyToMessage)
		if err != nil {
			telegram.ReplyMessage(bot, message, "请回复一条频道的图片消息")
			return
		}
		hash, err := common.GetPhash(fileBytes)
		if err != nil {
			telegram.ReplyMessage(bot, message, "请回复一条频道的图片消息")
			return
		}
		pictures, err := service.GetPicturesByHashHammingDistance(ctx, hash, 10)
		if err != nil || len(pictures) == 0 {
			telegram.ReplyMessage(bot, message, "请回复一条频道的图片消息")
			return
		}
		picture := pictures[0]
		_, err = telegram.SendPictureFileByMessageID(ctx, bot, message, picture.TelegramInfo.MessageID)
		if err != nil {
			telegram.ReplyMessage(bot, message, "文件发送失败: "+err.Error())
			return
		}
		return
	}
	pictureMessageID := messageOrigin.MessageID
	cmd, _, args := telegoutil.ParseCommand(message.Text)
	multiple := cmd == "files"
	artwork, err := service.GetArtworkByMessageID(ctx, pictureMessageID)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			telegram.ReplyMessage(bot, message, "这张图片未在数据库中呢")
			return
		}
		Logger.Errorf("获取作品失败: %s", err)
		telegram.ReplyMessage(bot, message, "获取失败, 去找管理员反馈吧~")
		return
	}

	if multiple {
		getArtworkFiles(ctx, bot, message, artwork)
		return
	}

	if len(args) > 0 {
		index, err := strconv.Atoi(args[0])
		if err == nil && index > 0 {
			if index > len(artwork.Pictures) {
				telegram.ReplyMessage(bot, message, "这个作品没有这么多图片")
				return
			}
			picture := artwork.Pictures[index-1]
			pictureMessageID = picture.TelegramInfo.MessageID
		}
	}
	_, err = telegram.SendPictureFileByMessageID(ctx, bot, message, pictureMessageID)
	if err != nil {
		telegram.ReplyMessage(bot, message, "文件发送失败: "+err.Error())
		return
	}
}

func getArtworkFiles(ctx context.Context, bot *telego.Bot, message telego.Message, artwork *types.Artwork) {
	defer func() {
		if r := recover(); r != nil {
			Logger.Fatalf("获取文件失败: %s", r)
		}
	}()
	for i, picture := range artwork.Pictures {
		var file telego.InputFile
		alreadyCached := picture.TelegramInfo.DocumentFileID != ""
		if alreadyCached {
			file = telegoutil.FileFromID(picture.TelegramInfo.DocumentFileID)
		} else {
			downloadingMessage, _ := telegram.ReplyMessage(bot, message, "正在下载第"+strconv.Itoa(i+1)+"张图片...")
			data, err := storage.GetStorage().GetFile(picture.StorageInfo)
			if err != nil {
				telegram.ReplyMessage(bot, message, "获取文件失败: "+err.Error())
				bot.DeleteMessage(telegoutil.Delete(message.Chat.ChatID(), downloadingMessage.MessageID))
				return
			}
			file = telegoutil.File(telegoutil.NameReader(bytes.NewReader(data), sources.GetFileName(artwork, picture)))
			bot.DeleteMessage(telegoutil.Delete(message.Chat.ChatID(), downloadingMessage.MessageID))
		}
		documentMessage, err := bot.SendDocument(telegoutil.Document(message.Chat.ChatID(), file).
			WithReplyParameters(&telego.ReplyParameters{
				MessageID: message.MessageID,
			}).WithCaption(artwork.Title + "_" + strconv.Itoa(i+1)))
		if err != nil {
			Logger.Errorf("发送文件失败: %s", err)
			bot.SendMessage(telegoutil.Messagef(
				message.Chat.ChatID(),
				"发送第 %d 张图片时失败",
				i+1,
			).WithReplyParameters(&telego.ReplyParameters{
				MessageID: message.MessageID,
			}))
			continue
		}
		if documentMessage != nil {
			picture.TelegramInfo.DocumentFileID = documentMessage.Document.FileID
			if service.UpdatePictureTelegramInfo(ctx, picture, picture.TelegramInfo) != nil {
				Logger.Warnf("更新图片信息失败: %s", err)
			}
			if alreadyCached {
				time.Sleep(time.Duration(config.Cfg.Telegram.Sleep) * time.Second)
			}
		}
	}
}

func randomPicture(ctx context.Context, bot *telego.Bot, message telego.Message) {
	cmd, _, args := telegoutil.ParseCommand(message.Text)
	argText := strings.ReplaceAll(strings.Join(args, " "), "\\", "")
	textArray, err := common.ParseStringTo2DArray(argText, "|", " ")
	if err != nil {
		Logger.Warnf("解析参数失败: %s", err)
		telegram.ReplyMessage(bot, message, "解析参数失败: "+err.Error()+"\n请使用'|'分隔'或'关系的关键词, 使用空格分隔'与'关系的关键词")
		return
	}
	r18 := cmd == "setu"
	r18Type := types.R18TypeNone
	if r18 {
		r18Type = types.R18TypeOnly
	}
	artwork, err := service.QueryArtworksByTexts(ctx, textArray, r18Type, 1)
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
		telegoutil.InlineKeyboard(telegram.GetPostedPictureInlineKeyboardButton(picture)),
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
	hasPermission := CheckPermissionInGroup(ctx, message, types.PermissionGetArtworkInfo)
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
		time.Sleep(1 * time.Second)
		if waitMessageID != 0 {
			bot.DeleteMessage(telegoutil.Delete(message.Chat.ChatID(), waitMessageID))
		}
	}()

	isAlreadyPosted := true
	artwork, err := service.GetArtworkByURL(ctx, sourceURL)
	if err != nil {
		cachedArtwork, err := service.GetCachedArtworkByURL(ctx, sourceURL)
		if err != nil && !hasPermission {
			return
		}
		isAlreadyPosted = false
		if cachedArtwork == nil {
			artwork, err = sources.GetArtworkInfo(sourceURL)
			if err != nil {
				telegram.ReplyMessage(bot, message, "获取作品信息失败: "+err.Error())
				return
			}
			if err := service.CreateCachedArtwork(ctx, artwork, types.ArtworkStatusCached); err != nil {
				Logger.Warnf("缓存作品失败: %s", err)
			}
		} else {
			artwork = cachedArtwork.Artwork
		}
	}

	var inputFile telego.InputFile
	if artwork.Pictures[0].TelegramInfo == nil || artwork.Pictures[0].TelegramInfo.PhotoFileID == "" {
		fileBytes, err := common.DownloadWithCache(artwork.Pictures[0].Original, nil)
		if err != nil {
			telegram.ReplyMessage(bot, message, "下载图片失败: "+err.Error())
			return
		}
		fileBytes, err = common.CompressImage(fileBytes, 5, 2560)
		if err != nil {
			telegram.ReplyMessage(bot, message, "压缩图片失败: "+err.Error())
			return
		}
		inputFile = telegoutil.File(telegoutil.NameReader(bytes.NewReader(fileBytes), artwork.Title))
	} else {
		inputFile = telegoutil.FileFromID(artwork.Pictures[0].TelegramInfo.PhotoFileID)
	}

	photo := telegoutil.Photo(message.Chat.ChatID(), inputFile).
		WithReplyParameters(&telego.ReplyParameters{MessageID: message.MessageID}).
		WithParseMode(telego.ModeHTML)

	deletedModel, _ := service.GetDeletedByURL(ctx, sourceURL)
	artworkInfoCaption := telegram.GetArtworkHTMLCaption(artwork)
	if deletedModel != nil && hasPermission {
		photo.WithCaption(artworkInfoCaption + fmt.Sprintf("\n\n这是一个在 %s 删除的作品\n如果发布则会取消删除", deletedModel.DeletedAt.Time().Format("2006-01-02 15:04:05")))
	} else {
		artworkInfoCaption += fmt.Sprintf("\n该作品共有%d张图片", len(artwork.Pictures))
		photo.WithCaption(artworkInfoCaption)
	}
	if isAlreadyPosted {
		photo.WithReplyMarkup(telegram.GetPostedPictureReplyMarkup(artwork.Pictures[0]))
	} else if hasPermission {
		id, err := service.CreateCallbackData(ctx, artwork.SourceURL)
		if err != nil {
			telegram.ReplyMessage(bot, message, "创建回调数据失败: "+err.Error())
			return
		}
		photo.WithReplyMarkup(telegoutil.InlineKeyboard(
			[]telego.InlineKeyboardButton{
				telegoutil.InlineKeyboardButton("发布").WithCallbackData("post_artwork " + id),
				telegoutil.InlineKeyboardButton("设为R18并发布").WithCallbackData("post_artwork_r18 " + id),
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
	hasPermission := CheckPermissionInGroup(ctx, message, types.PermissionSearchPicture)
	if message.ReplyToMessage == nil {
		telegram.ReplyMessage(bot, message, "请使用该命令回复一条图片消息")
		return
	}
	go telegram.ReplyMessage(bot, message, "少女祈祷中...")

	fileBytes, err := telegram.GetMessagePhotoFileBytes(bot, message.ReplyToMessage)
	if err != nil {
		telegram.ReplyMessage(bot, message, "获取图片文件失败: "+err.Error())
		return
	}
	hash, err := common.GetPhash(fileBytes)
	if err != nil {
		telegram.ReplyMessage(bot, message, "获取图片哈希失败: "+err.Error())
		return
	}
	pictures, err := service.GetPicturesByHashHammingDistance(ctx, hash, 10)
	if err != nil {
		telegram.ReplyMessage(bot, message, "搜索图片失败: "+err.Error())
		return
	}
	if len(pictures) > 0 {
		text := fmt.Sprintf("找到%d张相似或相同的图片\n\n", len(pictures))
		for _, picture := range pictures {
			artwork, err := service.GetArtworkByMessageID(ctx, picture.TelegramInfo.MessageID)
			if err != nil {
				text += common.EscapeMarkdown(fmt.Sprintf("%s 模糊度: %.2f\n\n", telegram.GetArtworkPostMessageURL(picture.TelegramInfo.MessageID), picture.BlurScore))
			}
			text += fmt.Sprintf("[%s\\_%d](%s)  ",
				common.EscapeMarkdown(artwork.Title),
				picture.Index+1,
				common.EscapeMarkdown(telegram.GetArtworkPostMessageURL(picture.TelegramInfo.MessageID)))
			text += common.EscapeMarkdown(fmt.Sprintf("模糊度: %.2f\n\n", picture.BlurScore))
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
	texts, err := common.ParseStringTo2DArray(queryText, "|", " ")
	if err != nil {
		Logger.Warnf("解析参数失败: %s", err)
		bot.AnswerInlineQuery(telegoutil.InlineQuery(query.ID, telegoutil.ResultArticle(uuid.NewString(), "解析参数失败", telegoutil.TextMessage("/setu"))))
		return
	}
	artworks, err := service.QueryArtworksByTexts(ctx, texts, types.R18TypeAll, 20)
	if err != nil || len(artworks) == 0 {
		Logger.Warnf("获取图片失败: %s", err)
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
		result.WithReplyMarkup(telegoutil.InlineKeyboard(telegram.GetPostedPictureInlineKeyboardButton(picture)))
		results = append(results, result)
	}
	if err := bot.AnswerInlineQuery(telegoutil.InlineQuery(query.ID, results...).WithCacheTime(1)); err != nil {
		Logger.Errorf("回复查询失败: %s", err)
	}
}

func calculatePicture(ctx context.Context, bot *telego.Bot, message telego.Message) {
	if message.ReplyToMessage == nil {
		telegram.ReplyMessage(bot, message, "请使用该命令回复一条图片消息")
		return
	}
	var waitMessageID int
	msg, err := telegram.ReplyMessage(bot, message, "少女做高数中...(((φ(◎ロ◎;)φ)))")
	if err == nil {
		waitMessageID = msg.MessageID
	}
	fileBytes, err := telegram.GetMessagePhotoFileBytes(bot, message.ReplyToMessage)
	if err != nil {
		telegram.ReplyMessage(bot, message, "获取图片文件失败: "+err.Error())
		return
	}
	hash, err := common.GetPhash(fileBytes)
	if err != nil {
		telegram.ReplyMessage(bot, message, "计算图片信息失败: "+err.Error())
		return
	}
	blurScore, err := common.GetBlurScore(fileBytes)
	if err != nil {
		telegram.ReplyMessage(bot, message, "计算图片信息失败: "+err.Error())
		return
	}
	width, height, err := common.GetImageSize(fileBytes)
	if err != nil {
		telegram.ReplyMessage(bot, message, "计算图片信息失败: "+err.Error())
		return
	}
	text := fmt.Sprintf(
		"<b>Hash</b>: <code>%s</code>\n<b>模糊度</b>: %.2f\n<b>尺寸</b>: %d x %d",
		hash, blurScore, width, height,
	)
	if waitMessageID == 0 {
		telegram.ReplyMessageWithHTML(bot, message, text)
		return
	}
	bot.EditMessageText(&telego.EditMessageTextParams{
		ChatID:    message.Chat.ChatID(),
		MessageID: waitMessageID,
		Text:      text,
		ParseMode: telego.ModeHTML,
	})

}
