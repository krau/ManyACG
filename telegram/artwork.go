package telegram

import (
	"ManyACG-Bot/common"
	"ManyACG-Bot/errors"
	"ManyACG-Bot/types"

	. "ManyACG-Bot/logger"

	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegoutil"
)

func PostArtwork(bot *telego.Bot, artwork *types.Artwork) (messages []telego.Message, err error) {
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
		photoURL := picture.Original
		if artwork.SourceType == types.SourceTypePixiv {
			photoURL = common.GetPixivRegularURL(photoURL)
		}
		photo := telegoutil.MediaPhoto(telegoutil.FileFromURL(photoURL))
		if i == 0 {
			photo = photo.WithCaption(GetArtworkMarkdownCaption(artwork)).WithParseMode(telego.ModeMarkdownV2)
		}
		if artwork.R18 {
			photo = photo.WithHasSpoiler()
		}
		inputMediaPhotos[i] = photo
	}

	return bot.SendMediaGroup(
		telegoutil.MediaGroup(
			ChannelChatID,
			inputMediaPhotos...,
		),
	)
}
