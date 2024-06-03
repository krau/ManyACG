package telegram

import (
	"ManyACG/common"
	"ManyACG/config"
	"ManyACG/errors"
	"ManyACG/storage"
	"ManyACG/types"
	"bytes"
	"time"

	. "ManyACG/logger"

	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegoutil"
)

func PostArtwork(bot *telego.Bot, artwork *types.Artwork, storage storage.Storage) ([]telego.Message, error) {
	if bot == nil {
		Logger.Fatal("Bot is nil")
		return nil, errors.ErrNilBot
	}
	if artwork == nil {
		Logger.Fatal("Artwork is nil")
		return nil, errors.ErrNilArtwork
	}

	inputMediaPhotos := make([]telego.InputMedia, len(artwork.Pictures))
	for i, picture := range artwork.Pictures {
		fileBytes, err := storage.GetFile(picture.StorageInfo)
		if err != nil {
			Logger.Errorf("failed to get file: %s", err)
			return nil, err
		}
		fileBytes, err = common.CompressImage(fileBytes, 10, 2560)
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
		inputMediaPhotos[i] = photo
	}

	if len(inputMediaPhotos) <= 10 {
		return bot.SendMediaGroup(
			telegoutil.MediaGroup(
				ChannelChatID,
				inputMediaPhotos...,
			),
		)
	}

	allMessages := make([]telego.Message, len(inputMediaPhotos))
	for i := 0; i < len(inputMediaPhotos); i += 10 {
		end := i + 10
		if end > len(inputMediaPhotos) {
			end = len(inputMediaPhotos)
		}
		mediaGroup := telegoutil.MediaGroup(
			ChannelChatID,
			inputMediaPhotos[i:end]...,
		)
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
		time.Sleep(time.Duration(int(config.Cfg.Telegram.Sleep)*len(inputMediaPhotos[i:end])) * time.Second)
	}
	return allMessages, nil
}
