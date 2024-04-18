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
		regularURL := strings.Replace(picture.Original, "img-original", "img-master", 1)
		regularURL = strings.Replace(regularURL, ".jpg", "_master1200.jpg", 1)
		regularURL = strings.Replace(regularURL, ".png", "_master1200.jpg", 1)
		photo := telegoutil.MediaPhoto(telegoutil.FileFromURL(regularURL))
		if i == 0 {
			caption := fmt.Sprintf("[*%s*](%s)", escapeMarkdown(artwork.Title), artwork.SourceURL)
			caption += "\n\n" + "*Author:* " + escapeMarkdown(artwork.Artist.Name)
			if artwork.Description != "" {
				if len(artwork.Description) > 233 {
					caption += "\n\n" + "_" + escapeMarkdown(artwork.Description[:233]) + "\\.\\.\\._"
				} else {
					caption += "\n\n" + "_" + escapeMarkdown(artwork.Description) + "_"
				}
			}
			tags := ""
			for _, tag := range artwork.Tags {
				tag = replaceChars(tag, []string{":", "：", "-", "（", "）", "「", "」", "*"}, "_")
				tag = replaceChars(tag, []string{"?"}, "")
				tag = replaceChars(tag, []string{"/"}, " #")
				tags += "\\#" + strings.Join(strings.Split(escapeMarkdown(tag), " "), "") + " "
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
			ChatID,
			inputMediaPhotos...,
		),
	)
}
