package telegram

import (
	"ManyACG-Bot/errors"
	"ManyACG-Bot/types"
	"fmt"
	"strings"

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
			photoURL = strings.Replace(photoURL, "img-original", "img-master", 1)
			photoURL = strings.Replace(photoURL, ".jpg", "_master1200.jpg", 1)
			photoURL = strings.Replace(photoURL, ".png", "_master1200.jpg", 1)
		}
		photo := telegoutil.MediaPhoto(telegoutil.FileFromURL(photoURL))
		if i == 0 {
			caption := fmt.Sprintf("[*%s*](%s)", (artwork.Title), artwork.SourceURL)
			caption += "\n\n" + "*Author:* " + EscapeMarkdown(artwork.Artist.Name)
			if artwork.Description != "" {
				if len(artwork.Description) > 233 {
					caption += "\n\n" + "_" + EscapeMarkdown(artwork.Description[:233]) + "\\.\\.\\._"
				} else {
					caption += "\n\n" + "_" + EscapeMarkdown(artwork.Description) + "_"
				}
			}
			tags := ""
			for _, tag := range artwork.Tags {
				tag = ReplaceChars(tag, []string{":", "：", "-", "（", "）", "「", "」", "*"}, "_")
				tag = ReplaceChars(tag, []string{"?"}, "")
				tag = ReplaceChars(tag, []string{"/"}, " "+"#")
				tags += "\\#" + strings.Join(strings.Split(EscapeMarkdown(tag), " "), "") + " "
			}
			caption += "\n\n" + tags
			photo = photo.WithCaption(caption).WithParseMode(telego.ModeMarkdownV2)
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
