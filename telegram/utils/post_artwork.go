package utils

import (
	"ManyACG/common"
	"ManyACG/config"
	"ManyACG/errors"
	. "ManyACG/logger"
	"ManyACG/service"
	"ManyACG/storage"
	"ManyACG/types"
	"bytes"
	"context"
	es "errors"
	"fmt"
	"runtime"
	"time"

	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegoutil"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func PostArtwork(ctx context.Context, bot *telego.Bot, artwork *types.Artwork) ([]telego.Message, error) {
	if bot == nil {
		return nil, errors.ErrNilBot
	}
	if artwork == nil {
		Logger.Fatal("Artwork is nil")
		return nil, errors.ErrNilArtwork
	}
	if len(artwork.Pictures) <= 10 {
		inputMediaPhotos, err := getInputMediaPhotos(ctx, artwork, 0, len(artwork.Pictures))
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
		inputMediaPhotos, err := getInputMediaPhotos(ctx, artwork, i, end)
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
func getInputMediaPhotos(ctx context.Context, artwork *types.Artwork, start, end int) ([]telego.InputMedia, error) {
	inputMediaPhotos := make([]telego.InputMedia, end-start)
	for i := start; i < end; i++ {
		picture := artwork.Pictures[i]
		fileBytes := common.GetReqCachedFile(picture.Original)
		if fileBytes == nil {
			var err error
			fileBytes, err = storage.GetFile(ctx, picture.StorageInfo.Original)
			if err != nil {
				Logger.Errorf("failed to get file: %s", err)
				return nil, err
			}
		}
		fileBytes, err := common.CompressImageToJPEG(fileBytes, 10, 2560, picture.Original)
		if err != nil {
			Logger.Errorf("failed to compress image: %s", err)
			return nil, err
		}
		photo := telegoutil.MediaPhoto(telegoutil.File(telegoutil.NameReader(bytes.NewReader(fileBytes), picture.StorageInfo.Original.Path)))
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

func PostAndCreateArtwork(ctx context.Context, artwork *types.Artwork, bot *telego.Bot, fromID int64, messageID int) error {
	artworkInDB, err := service.GetArtworkByURL(ctx, artwork.SourceURL)
	if err == nil && artworkInDB != nil {
		Logger.Debugf("Artwork %s already exists", artwork.Title)
		return fmt.Errorf("post artwork %s error: %w", artwork.Title, errors.ErrArtworkAlreadyExist)
	}
	if err != nil && !es.Is(err, mongo.ErrNoDocuments) {
		return err
	}
	if service.CheckDeletedByURL(ctx, artwork.SourceURL) {
		Logger.Debugf("Artwork %s is deleted", artwork.Title)
		return fmt.Errorf("post artwork %s error: %w", artwork.Title, errors.ErrArtworkDeleted)
	}
	showProgress := fromID != 0 && messageID != 0 && bot != nil
	if showProgress {
		defer bot.EditMessageReplyMarkup(&telego.EditMessageReplyMarkupParams{
			ChatID:      telegoutil.ID(fromID),
			MessageID:   messageID,
			ReplyMarkup: nil,
		})
	}
	for i, picture := range artwork.Pictures {
		if showProgress {
			go bot.EditMessageReplyMarkup(&telego.EditMessageReplyMarkupParams{
				ChatID:    telegoutil.ID(fromID),
				MessageID: messageID,
				ReplyMarkup: telegoutil.InlineKeyboard(
					[]telego.InlineKeyboardButton{
						telegoutil.InlineKeyboardButton(fmt.Sprintf("正在保存图片 %d/%d", i+1, len(artwork.Pictures))).WithCallbackData("noop"),
					},
				),
			})
		}
		info, err := storage.SaveAll(ctx, artwork, picture)
		if err != nil {
			Logger.Errorf("saving picture %d of artwork %s: %s", i, artwork.Title, err)
			return fmt.Errorf("saving picture %d of artwork %s: %w", i, artwork.Title, err)
		}
		artwork.Pictures[i].StorageInfo = info
	}
	if showProgress {
		go bot.EditMessageReplyMarkup(&telego.EditMessageReplyMarkupParams{
			ChatID:    telegoutil.ID(fromID),
			MessageID: messageID,
			ReplyMarkup: telegoutil.InlineKeyboard(
				[]telego.InlineKeyboardButton{
					telegoutil.InlineKeyboardButton("图片保存完成, 正在发布到频道...").WithCallbackData("noop"),
				},
			),
		})
	}
	messages, err := PostArtwork(ctx, bot, artwork)
	if err != nil {
		return fmt.Errorf("posting artwork [%s](%s): %w", artwork.Title, artwork.SourceURL, err)
	}
	Logger.Infof("Posted artwork %s", artwork.Title)
	if showProgress {
		go bot.EditMessageReplyMarkup(&telego.EditMessageReplyMarkupParams{
			ChatID:    telegoutil.ID(fromID),
			MessageID: messageID,
			ReplyMarkup: telegoutil.InlineKeyboard(
				[]telego.InlineKeyboardButton{
					telegoutil.InlineKeyboardButton("发布完成").WithCallbackData("noop"),
				},
			),
		})
	}
	for i, picture := range artwork.Pictures {
		if picture.TelegramInfo == nil {
			picture.TelegramInfo = &types.TelegramInfo{}
		}
		picture.TelegramInfo.MessageID = messages[i].MessageID
		picture.TelegramInfo.MediaGroupID = messages[i].MediaGroupID
		if messages[i].Photo != nil {
			picture.TelegramInfo.PhotoFileID = messages[i].Photo[len(messages[i].Photo)-1].FileID
		}
		if messages[i].Document != nil {
			picture.TelegramInfo.DocumentFileID = messages[i].Document.FileID
		}
	}

	_, err = service.CreateArtwork(ctx, artwork)
	if err != nil {
		go func() {
			if bot.DeleteMessages(&telego.DeleteMessagesParams{
				ChatID:     ChannelChatID,
				MessageIDs: GetMessageIDs(messages),
			}) != nil {
				Logger.Errorf("deleting messages: %s", err)
			}
		}()
		return fmt.Errorf("error when creating artwork %s: %w", artwork.SourceURL, err)
	}
	go afterCreate(context.TODO(), artwork, bot, fromID, messageID)
	return nil
}

func afterCreate(ctx context.Context, artwork *types.Artwork, bot *telego.Bot, fromID int64, _ int) {
	for _, picture := range artwork.Pictures {
		if err := service.ProcessPictureAndUpdate(ctx, picture); err != nil {
			Logger.Warnf("error when processing %d of artwork %s: %s", picture.Index, artwork.Title, err)
		}
	}
	runtime.GC()
	sendNotify := fromID != 0 && bot != nil
	artworkID, err := service.GetArtworkIDByPicture(ctx, artwork.Pictures[0])
	artworkTitleMarkdown := common.EscapeMarkdown(artwork.Title)
	if err != nil {
		Logger.Errorf("error when getting artwork by URL: %s", err)
		if sendNotify {
			bot.SendMessage(telegoutil.Messagef(telegoutil.ID(fromID),
				"刚刚发布的作品 [%s](%s) 后续处理失败\\: \n无法获取作品信息\\: %s", artworkTitleMarkdown, func() string {
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
		picture, err = service.GetPictureByMessageID(ctx, picture.TelegramInfo.MessageID)
		if err != nil {
			Logger.Warnf("error when getting picture by message ID: %s", err)
			continue
		}
		resPictures, err := service.GetPicturesByHashHammingDistance(ctx, picture.Hash, 10)
		if err != nil {
			Logger.Warnf("error when getting pictures by hash: %s", err)
			continue
		}
		similarPictures := make([]*types.Picture, 0)
		for _, resPicture := range resPictures {
			resArtworkID, err := service.GetArtworkIDByPicture(ctx, resPicture)
			if err != nil {
				Logger.Warnf("error when getting artwork ID by picture: %s", err)
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
		Logger.Noticef("Found %d similar pictures for %s", len(similarPictures), picture.Original)
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
		text += common.EscapeMarkdown(fmt.Sprintf("该图像模糊度: %.2f\n搜索到的相似图片列表:\n\n", picture.BlurScore))
		for _, similarPicture := range similarPictures {
			pictureObjectID, err := primitive.ObjectIDFromHex(similarPicture.ID)
			if err != nil {
				Logger.Errorf("invalid ObjectID: %s", similarPicture.ID)
				continue
			}

			artworkOfSimilarPicture, err := service.GetArtworkByID(ctx, pictureObjectID)
			if err != nil {
				Logger.Warnf("error when getting artwork by ID: %s", err)
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
			text += common.EscapeMarkdown(fmt.Sprintf("模糊度: %.2f\n\n", similarPicture.BlurScore))
		}
		text += "_模糊度使用原图文件计算得出, 越小图像质量越好_"
		_, err = bot.SendMessage(telegoutil.Messagef(telegoutil.ID(fromID), text).WithParseMode(telego.ModeMarkdownV2))
		if err != nil {
			Logger.Errorf("error when sending similar pictures: %s", err)
		}
	}
}
