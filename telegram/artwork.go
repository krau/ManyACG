package telegram

import (
	"ManyACG-Bot/errors"
	"ManyACG-Bot/model"
	"fmt"
	"strings"

	. "ManyACG-Bot/logger"

	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegoutil"
)

func PostArtwork(bot *telego.Bot, artwork *model.Artwork) (messages []telego.Message, err error) {
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
		photo := telegoutil.MediaPhoto(telegoutil.FileFromURL(picture.DirectURL))
		if i == 0 {
			caption := fmt.Sprintf("[*%s*](%s)", escapeMarkdown(artwork.Title), artwork.SourceURL)
			caption += "\n\n" + "Author: " + escapeMarkdown(artwork.Author)
			caption += "\n\n" + "Source: " + escapeMarkdown(artwork.SourceType)
			caption += "\n\n" + "Description: " + escapeMarkdown(artwork.Description)
			tags := ""
			for _, tag := range artwork.Tags {
				tag = replaceChars(tag, []string{":", "：", "-", "（", "）", "「", "」"}, "_")
				tags += "\\#" + strings.Join(strings.Split(escapeMarkdown(tag), " "), "") + " "
			}
			caption += "\n\n" + "Tags:" + tags
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
