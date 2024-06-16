package telegram

import (
	"ManyACG/common"
	"ManyACG/config"
	"ManyACG/errors"
	"ManyACG/storage"
	"ManyACG/types"
	"bytes"
	"runtime"
	"time"

	. "ManyACG/logger"

	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegoutil"
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
		time.Sleep(time.Duration(int(config.Cfg.Telegram.Sleep)*len(inputMediaPhotos)) * time.Second)
	}
	return allMessages, nil
}

// start from 0
func getInputMediaPhotos(artwork *types.Artwork, start, end int) ([]telego.InputMedia, error) {
	inputMediaPhotos := make([]telego.InputMedia, end-start)
	for i := start; i < end; i++ {
		picture := artwork.Pictures[i]
		fileBytes, err := storage.GetStorage().GetFile(picture.StorageInfo)
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
		inputMediaPhotos[i-start] = photo
		fileBytes = nil
	}
	runtime.GC()
	return inputMediaPhotos, nil
}
