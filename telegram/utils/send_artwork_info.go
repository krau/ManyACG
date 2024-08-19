package utils

import (
	"ManyACG/common"
	"ManyACG/errors"
	. "ManyACG/logger"
	"ManyACG/service"
	"ManyACG/sources"
	"ManyACG/types"
	"bytes"
	"context"
	es "errors"
	"fmt"

	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegoutil"
	"go.mongodb.org/mongo-driver/mongo"
)

type SendArtworkInfoParams struct {
	ChatID        *telego.ChatID
	SourceURL     string
	AppendCaption string
	Verify        bool
	HasPermission bool
	IgnoreDeleted bool
	ReplyParams   *telego.ReplyParameters
}

func SendArtworkInfo(ctx context.Context, bot *telego.Bot, params *SendArtworkInfoParams) error {
	if params.Verify {
		originSourceURL := params.SourceURL
		params.SourceURL = sources.FindSourceURL(params.SourceURL)
		if params.SourceURL == "" {
			return fmt.Errorf("无效的链接: %s", originSourceURL)
		}
	}
	deleteModel, _ := service.GetDeletedByURL(ctx, params.SourceURL)
	if deleteModel != nil && params.IgnoreDeleted {
		Logger.Debugf("已删除的作品: %s", params.SourceURL)
		return nil
	}
	isCreated := false
	artwork, err := service.GetArtworkByURL(ctx, params.SourceURL)
	if err != nil {
		if !params.HasPermission {
			return nil
		}
		if !es.Is(err, mongo.ErrNoDocuments) {
			return fmt.Errorf("获取作品信息失败: %w", err)
		}
		cachedArtwork, err := service.GetCachedArtworkByURLWithCache(ctx, params.SourceURL)
		if err != nil {
			return fmt.Errorf("缓存作品失败: %w", err)
		}
		artwork = cachedArtwork.Artwork
	} else {
		isCreated = true
	}

	caption := GetArtworkHTMLCaption(artwork) + fmt.Sprintf("\n<i>该作品共有%d张图片</i>", len(artwork.Pictures))
	if deleteModel != nil {
		caption += fmt.Sprintf("\n<i>这是一个在 %s 删除的作品\n如果发布则会取消删除</i>", common.EscapeHTML(deleteModel.DeletedAt.Time().Format("2006-01-02 15:04:05")))
	}

	replyMarkup, err := getArtworkInfoReplyMarkup(ctx, artwork, isCreated, params.HasPermission)
	if err != nil {
		return fmt.Errorf("获取 ReplyMarkup 失败: %w", err)
	}

	inputFile, needUpdatePreview, err := GetPicturePreviewInputFile(ctx, artwork.Pictures[0])
	if err != nil {
		return fmt.Errorf("获取预览图片失败: %w", err)
	}
	caption += fmt.Sprintf("\n%s", params.AppendCaption)
	if params.ChatID == nil {
		params.ChatID = &GroupChatID
	}
	photo := telegoutil.Photo(*params.ChatID, *inputFile).
		WithReplyParameters(params.ReplyParams).
		WithCaption(caption).
		WithReplyMarkup(replyMarkup).
		WithParseMode(telego.ModeHTML)

	if artwork.R18 && !needUpdatePreview {
		photo.WithHasSpoiler()
	}
	if bot == nil {
		return errors.ErrNilBot
	}
	msg, err := bot.SendPhoto(photo)
	if err != nil {
		return fmt.Errorf("发送图片失败: %w", err)
	}

	if !needUpdatePreview {
		cachedArtwork, err := service.GetCachedArtworkByURL(ctx, params.SourceURL)
		if err == nil {
			if cachedArtwork.Artwork.Pictures[0].TelegramInfo == nil {
				cachedArtwork.Artwork.Pictures[0].TelegramInfo = &types.TelegramInfo{}
			}
			cachedArtwork.Artwork.Pictures[0].TelegramInfo.PhotoFileID = msg.Photo[len(msg.Photo)-1].FileID
			if err := service.UpdateCachedArtwork(ctx, cachedArtwork); err != nil {
				Logger.Warnf("更新缓存作品失败: %s", err)
			}
		} else {
			Logger.Warnf("获取缓存作品失败: %s", err)
		}
		telegramInfo := artwork.Pictures[0].TelegramInfo
		if telegramInfo == nil {
			telegramInfo = &types.TelegramInfo{}
		}
		telegramInfo.PhotoFileID = msg.Photo[len(msg.Photo)-1].FileID
		if err := service.UpdatePictureTelegramInfo(ctx, artwork.Pictures[0], telegramInfo); err != nil {
			Logger.Warnf("更新图片信息失败: %s", err)
		}
		return nil
	}
	if err := updateLinkPreview(ctx, msg, artwork, bot, 0, photo); err != nil {
		Logger.Warnf("更新预览失败: %s", err)
		bot.EditMessageCaption(&telego.EditMessageCaptionParams{
			ChatID:      *params.ChatID,
			MessageID:   msg.MessageID,
			Caption:     caption + "\n<i>更新预览失败</i>",
			ParseMode:   telego.ModeHTML,
			ReplyMarkup: msg.ReplyMarkup,
		})
	}
	return nil
}

func getArtworkInfoReplyMarkup(ctx context.Context, artwork *types.Artwork, isCreated, hasPermission bool) (telego.ReplyMarkup, error) {
	if isCreated {
		baseKeyboard := GetPostedPictureInlineKeyboardButton(artwork, 0, ChannelChatID, BotUsername)
		if hasPermission {
			return telegoutil.InlineKeyboard(
				baseKeyboard,
				[]telego.InlineKeyboardButton{
					telegoutil.InlineKeyboardButton("更改R18").WithCallbackData("edit_artwork r18 " + artwork.ID + func() string {
						if artwork.R18 {
							return " 0"
						}
						return " 1"
					}()),
					telegoutil.InlineKeyboardButton("删除").WithCallbackData("delete_artwork " + artwork.ID),
				},
			), nil
		}
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
			telegoutil.InlineKeyboardButton("查重").WithCallbackData("search_picture " + cbId),
			telegoutil.InlineKeyboardButton("发布(R18)").WithCallbackData("post_artwork_r18 " + cbId),
		},
		previewKeyboard,
	), nil
}

func updateLinkPreview(ctx context.Context, targetMessage *telego.Message, artwork *types.Artwork, bot *telego.Bot, pictureIndex uint, photoParams *telego.SendPhotoParams) error {
	if pictureIndex >= uint(len(artwork.Pictures)) {
		return errors.ErrIndexOOB
	}
	var inputFile telego.InputFile
	fileBytes, err := common.DownloadWithCache(ctx, artwork.Pictures[pictureIndex].Original, nil)
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
	if err == nil {
		if cachedArtwork.Status == types.ArtworkStatusPosted {
			replyMarkup = GetPostedPictureReplyMarkup(artwork, pictureIndex, ChannelChatID, BotUsername)
		} else if cachedArtwork.Status == types.ArtworkStatusCached {
			replyMarkup = targetMessage.ReplyMarkup
		} else {
			mediaPhoto.WithCaption(photoParams.Caption + "\n<i>正在发布...</i>").WithParseMode(telego.ModeHTML)
		}
	} else {
		artwork, err := service.GetArtworkByURL(ctx, artwork.SourceURL)
		if err != nil {
			return err
		}
		replyMarkup = GetPostedPictureReplyMarkup(artwork, pictureIndex, ChannelChatID, BotUsername)
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
	if err := service.UpdateCachedArtwork(ctx, cachedArtwork); err != nil {
		Logger.Warnf("更新缓存作品失败: %s", err)
	}
	if err := service.UpdatePictureTelegramInfo(ctx, cachedArtwork.Artwork.Pictures[pictureIndex], cachedArtwork.Artwork.Pictures[pictureIndex].TelegramInfo); err != nil {
		Logger.Warnf("更新图片信息失败: %s", err)
	}
	return nil
}
