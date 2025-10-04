package utils

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/samber/oops"

	"github.com/google/uuid"
	"github.com/krau/ManyACG/common"
	"github.com/krau/ManyACG/common/imgtool"
	"github.com/krau/ManyACG/config"
	"github.com/krau/ManyACG/errs"
	"github.com/krau/ManyACG/internal/common/httpclient"
	"github.com/krau/ManyACG/internal/model/command"
	"github.com/krau/ManyACG/internal/model/entity"
	"github.com/krau/ManyACG/internal/shared"
	"gorm.io/datatypes"

	"github.com/krau/ManyACG/service"
	"github.com/krau/ManyACG/storage"

	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegoapi"
	"github.com/mymmrac/telego/telegoutil"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func SendArtworkMediaGroup(ctx context.Context, bot *telego.Bot, chatID telego.ChatID, artwork *entity.Artwork) ([]telego.Message, error) {
	if bot == nil {
		return nil, oops.New("bot is nil")
	}
	if artwork == nil {
		return nil, oops.New("artwork is nil")
	}
	if len(artwork.Pictures) <= 10 {
		inputMediaPhotos, err := GetArtworkInputMediaPhotos(ctx, artwork, 0, len(artwork.Pictures))
		if err != nil {
			return nil, err
		}
		return bot.SendMediaGroup(ctx,
			telegoutil.MediaGroup(
				chatID,
				inputMediaPhotos...,
			),
		)
	}
	allMessages := make([]telego.Message, len(artwork.Pictures))
	retryCount := 0
	maxRetry := config.Cfg.Telegram.Retry.MaxAttempts
	for i := 0; i < len(artwork.Pictures); i += 10 {
		end := i + 10
		if end > len(artwork.Pictures) {
			end = len(artwork.Pictures)
		}
		inputMediaPhotos, err := GetArtworkInputMediaPhotos(ctx, artwork, i, end)
		if err != nil {
			return nil, err
		}
		mediaGroup := telegoutil.MediaGroup(chatID, inputMediaPhotos...)
		if i > 0 {
			mediaGroup = mediaGroup.WithReplyParameters(&telego.ReplyParameters{
				ChatID:    chatID,
				MessageID: allMessages[i-1].MessageID,
			})
		}
		messages, err := bot.SendMediaGroup(ctx, mediaGroup)
		if err != nil {
			var apiError *telegoapi.Error
			if errors.As(err, &apiError) {
				switch apiError.ErrorCode {
				case 429:
					if apiError.Parameters == nil {
						return nil, err
					}
					if retryCount > maxRetry {
						return nil, fmt.Errorf("rate limited: %w", err)
					}
					retryAfter := apiError.Parameters.RetryAfter + (retryCount * int(config.Cfg.Telegram.Sleep))
					common.Logger.Warnf("Rate limited, retry after %d seconds", retryAfter)
					time.Sleep(time.Duration(retryAfter) * time.Second)
					i -= 10
					retryCount++
					continue
				default:
					return nil, apiError
				}
			} else if strings.Contains(err.Error(), "Too Many Requests") {
				// 偶尔会有无法 As 到 telegoapi.Error 的情况
				if retryCount > maxRetry {
					return nil, fmt.Errorf("rate limited: %w", err)
				}
				retryAfter := len(inputMediaPhotos) * int(config.Cfg.Telegram.Sleep)
				common.Logger.Warnf("Rate limited, retry after %d seconds", retryAfter)
				time.Sleep(time.Duration(retryAfter) * time.Second)
				i -= 10
				retryCount++
				continue
			}
			common.Logger.Errorf("failed to send media group: %s", err)
			return nil, err
		}
		copy(allMessages[i:], messages)
		retryCount = 0
	}
	return allMessages, nil
}

// start from 0
func GetArtworkInputMediaPhotos(ctx context.Context, artwork *entity.Artwork, start, end int) ([]telego.InputMedia, error) {
	inputMediaPhotos := make([]telego.InputMedia, end-start)
	for i := start; i < end; i++ {
		picture := artwork.Pictures[i]
		var photo *telego.InputMediaPhoto
		if id := picture.TelegramInfo.Data().PhotoFileID; id != "" {
			photo = telegoutil.MediaPhoto(telegoutil.FileFromID(id))
		}
		if photo == nil {
			var fileBytes []byte
			var err error
			fileBytes, err = common.GetReqCachedFile(picture.Original)
			if err != nil {
				if picture.StorageInfo.Data() == shared.ZeroStorageInfo || picture.StorageInfo.Data().Original == nil {
					fileBytes, err = common.DownloadWithCache(ctx, picture.Original, nil)
					if err != nil {
						common.Logger.Errorf("failed to download file: %s", err)
						return nil, err
					}
				} else {
					fileBytes, err = storage.GetFile(ctx, picture.StorageInfo.Data().Original)
					if err != nil {
						common.Logger.Errorf("failed to get file: %s", err)
						return nil, err
					}
				}
			}
			fileBytes, err = imgtool.CompressImageForTelegram(fileBytes)
			if err != nil {
				common.Logger.Errorf("failed to compress image: %s", err)
				return nil, err
			}
			photo = telegoutil.MediaPhoto(telegoutil.File(telegoutil.NameReader(bytes.NewReader(fileBytes), uuid.New().String())))
		}
		if i == 0 {
			photo = photo.WithCaption(GetArtworkHTMLCaption(artwork)).WithParseMode(telego.ModeHTML)
		}
		if artwork.R18 {
			photo = photo.WithHasSpoiler()
		}
		inputMediaPhotos[i-start] = photo
	}
	return inputMediaPhotos, nil
}

func PostAndCreateArtwork(ctx context.Context, artwork *command.ArtworkCreation, bot *telego.Bot, fromID int64, messageID int) error {
	artworkInDB, err := service.Default().GetArtworkByURL(ctx, artwork.SourceURL)
	if err == nil && artworkInDB != nil {
		common.Logger.Debugf("Artwork %s already exists", artwork.Title)
		return fmt.Errorf("post artwork %s error: %w", artwork.Title, errs.ErrArtworkAlreadyExist)
	}
	if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
		return err
	}
	if service.Default().CheckDeletedByURL(ctx, artwork.SourceURL) {
		common.Logger.Debugf("Artwork %s is deleted", artwork.Title)
		return fmt.Errorf("post artwork %s error: %w", artwork.Title, errs.ErrArtworkDeleted)
	}
	showProgress := fromID != 0 && messageID != 0 && bot != nil
	if showProgress {
		defer bot.EditMessageReplyMarkup(ctx, &telego.EditMessageReplyMarkupParams{
			ChatID:      telegoutil.ID(fromID),
			MessageID:   messageID,
			ReplyMarkup: nil,
		})
	}
	for i, picture := range artwork.Pictures {
		if showProgress {
			go bot.EditMessageReplyMarkup(ctx, &telego.EditMessageReplyMarkupParams{
				ChatID:    telegoutil.ID(fromID),
				MessageID: messageID,
				ReplyMarkup: telegoutil.InlineKeyboard(
					[]telego.InlineKeyboardButton{
						telegoutil.InlineKeyboardButton("正在处理存储...").WithCallbackData("noop")},
				),
			})
		}
		fileBytes, err := httpclient.DownloadWithCache(ctx, picture.Original, nil)
		info, err := service.Default().StorageSaveAll(ctx, artwork, picture)
		if err != nil {
			common.Logger.Errorf("saving picture %d of artwork %s: %s", i, artwork.Title, err)
			return fmt.Errorf("saving picture %d of artwork %s: %w", i, artwork.Title, err)
		}
		artwork.Pictures[i].StorageInfo = datatypes.NewJSONType(*info)
	}
	if showProgress {
		go bot.EditMessageReplyMarkup(ctx, &telego.EditMessageReplyMarkupParams{
			ChatID:    telegoutil.ID(fromID),
			MessageID: messageID,
			ReplyMarkup: telegoutil.InlineKeyboard(
				[]telego.InlineKeyboardButton{
					telegoutil.InlineKeyboardButton("正在发布到频道...").WithCallbackData("noop"),
				},
			),
		})
	}
	messages, err := SendArtworkMediaGroup(ctx, bot, ChannelChatID, artwork)
	if err != nil {
		return fmt.Errorf("posting artwork [%s](%s): %w", artwork.Title, artwork.SourceURL, err)
	}
	common.Logger.Infof("Posted artwork %s", artwork.Title)
	if showProgress {
		go bot.EditMessageReplyMarkup(ctx, &telego.EditMessageReplyMarkupParams{
			ChatID:    telegoutil.ID(fromID),
			MessageID: messageID,
			ReplyMarkup: telegoutil.InlineKeyboard(
				[]telego.InlineKeyboardButton{
					telegoutil.InlineKeyboardButton("发布完成").WithCallbackData("noop"),
				},
			),
		})
	}
	// [TODO] update picture telegram info
	// for i, picture := range artwork.Pictures {
	// 	picture.TelegramInfo.Data().MessageID = messages[i].MessageID
	// 	picture.TelegramInfo.Data().MediaGroupID = messages[i].MediaGroupID
	// 	if messages[i].Photo != nil {
	// 		picture.TelegramInfo.PhotoFileID = messages[i].Photo[len(messages[i].Photo)-1].FileID
	// 	}
	// 	if messages[i].Document != nil {
	// 		picture.TelegramInfo.DocumentFileID = messages[i].Document.FileID
	// 	}
	// }

	artwork, err = service.CreateArtwork(ctx, &command.ArtworkCreation{
		Title: artwork.Title, Description: artwork.Description,
		SourceType: artwork.SourceType, SourceURL: artwork.SourceURL,
		R18: artwork.R18,
		Artist: command.ArtworkArtistCreation{
			Name:     artwork.Artist.Name,
			UID:      artwork.Artist.UID,
			Username: artwork.Artist.Username,
		},
		Tags: artwork.TagsNames(),
	})
	if err != nil {
		go func() {
			if bot.DeleteMessages(ctx, &telego.DeleteMessagesParams{
				ChatID:     ChannelChatID,
				MessageIDs: GetMessageIDs(messages),
			}) != nil {
				common.Logger.Errorf("deleting messages: %s", err)
			}
		}()
		return fmt.Errorf("error when creating artwork: %w", err)
	}
	go afterCreate(context.TODO(), artwork, bot, fromID)
	return nil
}

func afterCreate(ctx context.Context, artwork *entity.Artwork, bot *telego.Bot, fromID int64) {
	time.Sleep(3 * time.Second) // 等待 Telegram 完全处理完消息
	objectID, err := primitive.ObjectIDFromHex(artwork.ID)
	if err != nil {
		common.Logger.Fatalf("invalid ObjectID: %s", artwork.ID)
		return
	}
	for _, picture := range artwork.Pictures {
		service.AddProcessPictureTask(ctx, picture)
	}
	checkDuplicate(ctx, artwork, bot, fromID)
	if common.TaggerClient != nil && config.Cfg.Tagger.TagNew {
		go service.AddPredictArtworkTagTask(ctx, objectID, TgService)
	} else {
		go recaptionArtwork(ctx, artwork, bot)
	}
}

func checkDuplicate(ctx context.Context, artwork *entity.Artwork, bot *telego.Bot, fromID int64) {
	sendNotify := fromID != 0 && bot != nil
	artworkID, err := primitive.ObjectIDFromHex(artwork.ID)
	artworkTitleMarkdown := common.EscapeMarkdown(artwork.Title)
	if err != nil {
		common.Logger.Errorf("invalid ObjectID: %s", artwork.ID)
		if sendNotify {
			bot.SendMessage(ctx, telegoutil.Messagef(telegoutil.ID(fromID),
				"刚刚发布的作品 [%s](%s) 后续处理失败\\: \n无效的ObjectID\\: %s", artworkTitleMarkdown, func() string {
					if artwork.Pictures[0].TelegramInfo.MessageID != 0 {
						return GetArtworkPostMessageURL(artwork.Pictures[0].TelegramInfo.MessageID, ChannelChatID)
					}
					return artwork.SourceURL
				}(), err).
				WithParseMode(telego.ModeMarkdownV2))
		}
		return
	}

	for _, picture := range artwork.Pictures {
		pictureID, err := primitive.ObjectIDFromHex(picture.ID)
		if err != nil {
			common.Logger.Errorf("invalid ObjectID: %s", picture.ID)
			continue
		}
		picture, err = service.GetPictureByID(ctx, pictureID)
		if err != nil {
			common.Logger.Warnf("error when getting picture by message ID: %s", err)
			continue
		}
		resPictures, err := service.GetPicturesByHashHammingDistance(ctx, picture.Hash, 10)
		if err != nil {
			common.Logger.Warnf("error when getting pictures by hash: %s", err)
			continue
		}
		similarPictures := make([]*entity.Picture, 0)
		for _, resPicture := range resPictures {
			resArtworkID, err := primitive.ObjectIDFromHex(resPicture.ArtworkID)
			if err != nil {
				common.Logger.Warnf("invalid ObjectID: %s", resPicture.ArtworkID)
				continue
			}
			if resArtworkID == artworkID {
				continue
			}
			similarPictures = append(similarPictures, resPicture)
		}
		if len(similarPictures) == 0 {
			continue
		}
		common.Logger.Noticef("Found %d similar pictures for %s", len(similarPictures), picture.Original)
		if !sendNotify {
			continue
		}

		text := fmt.Sprintf("*刚刚发布的作品 [%s](%s) 中第 %d 张图片搜索到有%d个相似图片*\n",
			artworkTitleMarkdown,
			common.EscapeMarkdown(func() string {
				if picture.TelegramInfo.MessageID != 0 {
					return GetArtworkPostMessageURL(picture.TelegramInfo.MessageID, ChannelChatID)
				}
				return artwork.SourceURL
			}()),
			picture.Index+1,
			len(similarPictures))
		text += common.EscapeMarkdown("搜索到的相似图片列表:\n\n")
		for _, similarPicture := range similarPictures {
			objectID, err := primitive.ObjectIDFromHex(similarPicture.ArtworkID)
			if err != nil {
				common.Logger.Errorf("invalid ObjectID: %s", similarPicture.ArtworkID)
				continue
			}

			artworkOfSimilarPicture, err := service.GetArtworkByID(ctx, objectID)
			if err != nil {
				common.Logger.Warnf("error when getting artwork by ID: %s", err)
				continue
			}
			text += fmt.Sprintf("[%s\\_%d](%s)  ",
				common.EscapeMarkdown(artworkOfSimilarPicture.Title),
				similarPicture.Index+1,
				common.EscapeMarkdown(func() string {
					if similarPicture.TelegramInfo.MessageID != 0 {
						return GetArtworkPostMessageURL(similarPicture.TelegramInfo.MessageID, ChannelChatID)
					}
					return artworkOfSimilarPicture.SourceURL
				}()))
		}
		_, err = bot.SendMessage(ctx, telegoutil.Message(telegoutil.ID(fromID), text).WithParseMode(telego.ModeMarkdownV2))
		if err != nil {
			common.Logger.Errorf("error when sending similar pictures: %s", err)
		}
	}
}

func recaptionArtwork(ctx context.Context, artwork *entity.Artwork, bot *telego.Bot) {
	newArtwork, err := service.GetArtworkByURL(ctx, artwork.SourceURL)
	if err != nil {
		common.Logger.Errorf("error when getting artwork by URL: %s", err)
		return
	}
	if newArtwork.Pictures[0].TelegramInfo == nil || newArtwork.Pictures[0].TelegramInfo.MessageID == 0 {
		return
	}
	newCaption := GetArtworkHTMLCaption(newArtwork)
	_, err = bot.EditMessageCaption(ctx, &telego.EditMessageCaptionParams{
		ChatID:    ChannelChatID,
		MessageID: newArtwork.Pictures[0].TelegramInfo.MessageID,
		Caption:   newCaption,
		ParseMode: telego.ModeHTML,
	})
	if err != nil {
		common.Logger.Warnf("error when recaption artwork: %s", err)
	}
}
