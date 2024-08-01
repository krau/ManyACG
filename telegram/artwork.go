package telegram

import (
	"ManyACG/common"
	"ManyACG/config"
	"ManyACG/errors"
	"ManyACG/service"
	"ManyACG/sources"
	"ManyACG/storage"
	"ManyACG/types"
	"bytes"
	"context"
	es "errors"
	"fmt"
	"runtime"
	"time"

	. "ManyACG/logger"

	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegoutil"
	"go.mongodb.org/mongo-driver/mongo"
)

func PostArtwork(bot *telego.Bot, artwork *types.Artwork) ([]telego.Message, error) {
	if bot == nil {
		Logger.Fatal("Bot is nil")
		return nil, errors.ErrNilBot
	}
	if artwork == nil {
		Logger.Fatal("Artwork is nil")
		return nil, errors.ErrNilArtwork
	}
	if len(artwork.Pictures) <= 10 {
		inputMediaPhotos, err := getInputMediaPhotos(artwork, 0, len(artwork.Pictures))
		if err != nil {
			return nil, err
		}
		return bot.SendMediaGroup(
			telegoutil.MediaGroup(
				ChannelChatID,
				inputMediaPhotos...,
			),
		)
	}
	allMessages := make([]telego.Message, len(artwork.Pictures))
	for i := 0; i < len(artwork.Pictures); i += 10 {
		end := i + 10
		if end > len(artwork.Pictures) {
			end = len(artwork.Pictures)
		}
		inputMediaPhotos, err := getInputMediaPhotos(artwork, i, end)
		if err != nil {
			return nil, err
		}
		mediaGroup := telegoutil.MediaGroup(ChannelChatID, inputMediaPhotos...)
		if i > 0 {
			mediaGroup = mediaGroup.WithReplyParameters(&telego.ReplyParameters{
				ChatID:    ChannelChatID,
				MessageID: allMessages[i-1].MessageID,
			})
		}
		messages, err := bot.SendMediaGroup(mediaGroup)
		if err != nil {
			return nil, err
		}
		copy(allMessages[i:], messages)
		if i+10 < len(artwork.Pictures) {
			time.Sleep(time.Duration(int(config.Cfg.Telegram.Sleep)*len(inputMediaPhotos)) * time.Second)
		}
	}
	return allMessages, nil
}

// start from 0
func getInputMediaPhotos(artwork *types.Artwork, start, end int) ([]telego.InputMedia, error) {
	inputMediaPhotos := make([]telego.InputMedia, end-start)
	for i := start; i < end; i++ {
		picture := artwork.Pictures[i]
		fileBytes := common.GetReqCachedFile(picture.Original)
		if fileBytes == nil {
			var err error
			fileBytes, err = storage.GetStorage().GetFile(picture.StorageInfo)
			if err != nil {
				Logger.Errorf("failed to get file: %s", err)
				return nil, err
			}
		}
		fileBytes, err := common.CompressImageWithCache(fileBytes, 10, 2560, picture.Original)
		if err != nil {
			Logger.Errorf("failed to compress image: %s", err)
			return nil, err
		}
		photo := telegoutil.MediaPhoto(telegoutil.File(telegoutil.NameReader(bytes.NewReader(fileBytes), picture.StorageInfo.Path)))
		if i == 0 {
			photo = photo.WithCaption(GetArtworkHTMLCaption(artwork)).WithParseMode(telego.ModeHTML)
		}
		if artwork.R18 {
			photo = photo.WithHasSpoiler()
		}
		inputMediaPhotos[i-start] = photo
		fileBytes = nil
	}
	runtime.GC()
	return inputMediaPhotos, nil
}

func SendArtworkInfo(ctx context.Context, bot *telego.Bot, sourceURL string, needVerifySourceURL bool, chatID *telego.ChatID,
	hasPermission bool, appendCaption string, ignoreDeleted bool, replyParameters *telego.ReplyParameters) error {
	if needVerifySourceURL {
		originSourceURL := sourceURL
		sourceURL = sources.FindSourceURL(sourceURL)
		if sourceURL == "" {
			return fmt.Errorf("无效的链接: %s", originSourceURL)
		}
	}
	deleteModel, _ := service.GetDeletedByURL(ctx, sourceURL)
	if deleteModel != nil && ignoreDeleted {
		Logger.Debugf("已删除的作品: %s", sourceURL)
		return nil
	}
	isCreated := false
	artwork, err := service.GetArtworkByURL(ctx, sourceURL)
	if err != nil {
		if !hasPermission {
			return nil
		}
		if !es.Is(err, mongo.ErrNoDocuments) {
			return fmt.Errorf("获取作品信息失败: %w", err)
		}
		cachedArtwork, err := service.GetCachedArtworkByURLWithCache(ctx, sourceURL)
		if err != nil {
			return fmt.Errorf("缓存作品失败: %w", err)
		}
		artwork = cachedArtwork.Artwork
		go downloadAndCompressArtwork(ctx, artwork)
	} else {
		isCreated = true
	}

	caption := GetArtworkHTMLCaption(artwork) + fmt.Sprintf("\n<i>该作品共有%d张图片</i>", len(artwork.Pictures))
	if deleteModel != nil {
		caption += fmt.Sprintf("\n<i>这是一个在 %s 删除的作品\n如果发布则会取消删除</i>", common.EscapeHTML(deleteModel.DeletedAt.Time().Format("2006-01-02 15:04:05")))
	}

	replyMarkup, err := getArtworkInfoReplyMarkup(ctx, artwork, isCreated)
	if err != nil {
		return fmt.Errorf("获取 ReplyMarkup 失败: %w", err)
	}

	inputFile, needUpdatePreview, err := GetPicturePreviewInputFile(ctx, artwork.Pictures[0])
	if err != nil {
		return fmt.Errorf("获取预览图片失败: %w", err)
	}
	caption += fmt.Sprintf("\n%s", appendCaption)
	if chatID == nil {
		chatID = &GroupChatID
	}
	photo := telegoutil.Photo(*chatID, *inputFile).
		WithReplyParameters(replyParameters).
		WithCaption(caption).
		WithReplyMarkup(replyMarkup).
		WithParseMode(telego.ModeHTML)

	if artwork.R18 && !needUpdatePreview {
		photo.WithHasSpoiler()
	}
	if bot == nil {
		bot = Bot
	}
	msg, err := bot.SendPhoto(photo)
	if err != nil {
		return fmt.Errorf("发送图片失败: %w", err)
	}

	if !needUpdatePreview {
		cachedArtwork, err := service.GetCachedArtworkByURL(ctx, sourceURL)
		if err != nil {
			Logger.Warnf("获取缓存作品失败: %s", err)
			return nil
		}
		if cachedArtwork.Artwork.Pictures[0].TelegramInfo == nil {
			cachedArtwork.Artwork.Pictures[0].TelegramInfo = &types.TelegramInfo{}
		}
		cachedArtwork.Artwork.Pictures[0].TelegramInfo.PhotoFileID = msg.Photo[len(msg.Photo)-1].FileID
		if err := service.UpdateCachedArtwork(ctx, cachedArtwork); err != nil {
			Logger.Warnf("更新缓存作品失败: %s", err)
		}
		return nil
	}
	if err := updateLinkPreview(ctx, msg, artwork, bot, 0, photo); err != nil {
		Logger.Warnf("更新预览失败: %s", err)
		bot.EditMessageCaption(&telego.EditMessageCaptionParams{
			ChatID:      *chatID,
			MessageID:   msg.MessageID,
			Caption:     caption + "\n<i>更新预览失败</i>",
			ParseMode:   telego.ModeHTML,
			ReplyMarkup: msg.ReplyMarkup,
		})
	}
	return nil
}

func getArtworkInfoReplyMarkup(ctx context.Context, artwork *types.Artwork, isCreated bool) (telego.ReplyMarkup, error) {
	if isCreated {
		return GetPostedPictureReplyMarkup(artwork.Pictures[0]), nil
	}
	cbId, err := service.CreateCallbackData(ctx, artwork.SourceURL)
	if err != nil {
		return nil, fmt.Errorf("创建回调数据失败: %w", err)
	}

	previewKeyboard := []telego.InlineKeyboardButton{}
	if len(artwork.Pictures) > 1 {
		previewKeyboard = append(previewKeyboard, telegoutil.InlineKeyboardButton(fmt.Sprintf("删除这张(%d)", 1)).WithCallbackData("artwork_preview "+cbId+" delete 0 0"))
		previewKeyboard = append(previewKeyboard, telegoutil.InlineKeyboardButton("下一张").WithCallbackData("artwork_preview "+cbId+" preview 1 0"))
	}
	return telegoutil.InlineKeyboard(
		[]telego.InlineKeyboardButton{
			telegoutil.InlineKeyboardButton("发布").WithCallbackData("post_artwork " + cbId),
			telegoutil.InlineKeyboardButton("设为R18并发布").WithCallbackData("post_artwork_r18 " + cbId),
		},
		previewKeyboard,
	), nil
}

func downloadAndCompressArtwork(ctx context.Context, artwork *types.Artwork) {
	for i, picture := range artwork.Pictures {
		if i == 0 {
			continue
		}
		if picture.TelegramInfo != nil && picture.TelegramInfo.PhotoFileID != "" {
			continue
		}
		cachedArtwork, err := service.GetCachedArtworkByURL(ctx, artwork.SourceURL)
		if err != nil {
			Logger.Warnf("获取缓存作品失败: %s", err)
			continue
		}
		if cachedArtwork.Status != types.ArtworkStatusCached {
			break
		}
		fileBytes, err := common.DownloadWithCache(picture.Original, nil)
		if err != nil {
			Logger.Warnf("下载图片失败: %s", err)
			continue
		}
		_, err = common.CompressImageWithCache(fileBytes, 10, 2560, picture.Original)
		if err != nil {
			Logger.Warnf("压缩图片失败: %s", err)
		}
	}
}

func updateLinkPreview(ctx context.Context, targetMessage *telego.Message, artwork *types.Artwork, bot *telego.Bot, pictureIndex uint, photoParams *telego.SendPhotoParams) error {
	if pictureIndex >= uint(len(artwork.Pictures)) {
		return errors.ErrIndexOOB
	}
	var inputFile telego.InputFile
	fileBytes, err := common.DownloadWithCache(artwork.Pictures[pictureIndex].Original, nil)
	if err != nil {
		return err
	}
	fileBytes, err = common.CompressImageWithCache(fileBytes, 10, 2560, artwork.Pictures[pictureIndex].Original)
	if err != nil {
		return err
	}
	inputFile = telegoutil.File(telegoutil.NameReader(bytes.NewReader(fileBytes), artwork.Title))
	mediaPhoto := telegoutil.MediaPhoto(inputFile)
	mediaPhoto.WithCaption(photoParams.Caption).WithParseMode(photoParams.ParseMode)

	var replyMarkup *telego.InlineKeyboardMarkup
	cachedArtwork, err := service.GetCachedArtworkByURL(ctx, artwork.SourceURL)
	if err != nil {
		return err
	}
	if cachedArtwork.Status == types.ArtworkStatusPosted {
		replyMarkup = GetPostedPictureReplyMarkup(artwork.Pictures[pictureIndex])
	} else if cachedArtwork.Status == types.ArtworkStatusCached {
		replyMarkup = targetMessage.ReplyMarkup
	} else {
		mediaPhoto.WithCaption(photoParams.Caption + "\n<i>正在发布...</i>").WithParseMode(telego.ModeHTML)
	}
	msg, err := bot.EditMessageMedia(&telego.EditMessageMediaParams{
		ChatID:      targetMessage.Chat.ChatID(),
		MessageID:   targetMessage.MessageID,
		Media:       mediaPhoto,
		ReplyMarkup: replyMarkup,
	})
	if err != nil {
		return err
	}
	if cachedArtwork.Artwork.Pictures[pictureIndex].TelegramInfo == nil {
		cachedArtwork.Artwork.Pictures[pictureIndex].TelegramInfo = &types.TelegramInfo{}
	}
	cachedArtwork.Artwork.Pictures[pictureIndex].TelegramInfo.PhotoFileID = msg.Photo[len(msg.Photo)-1].FileID
	service.UpdateCachedArtwork(ctx, cachedArtwork)
	return nil
}
