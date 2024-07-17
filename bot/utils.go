package bot

import (
	"ManyACG/common"
	"ManyACG/errors"
	"ManyACG/service"
	"ManyACG/sources"
	"ManyACG/telegram"
	"ManyACG/types"
	"bytes"
	"context"

	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegoutil"
)

func CheckPermissionInGroup(ctx context.Context, message telego.Message, permissions ...types.Permission) bool {
	chatID := message.Chat.ID
	if message.Chat.Type != telego.ChatTypeGroup && message.Chat.Type != telego.ChatTypeSupergroup {
		chatID = message.From.ID
	}
	if !service.CheckAdminPermission(ctx, chatID, permissions...) {
		return service.CheckAdminPermission(ctx, message.From.ID, permissions...)
	}
	return true
}

func CheckPermissionForQuery(ctx context.Context, query telego.CallbackQuery, permissions ...types.Permission) bool {
	if !service.CheckAdminPermission(ctx, query.From.ID, permissions...) &&
		!service.CheckAdminPermission(ctx, query.Message.GetChat().ID, permissions...) {
		return false
	}
	return true
}

func FindSourceURLForMessage(message *telego.Message) string {
	text := message.Text
	text += message.Caption + " "
	for _, entity := range message.Entities {
		if entity.Type == telego.EntityTypeTextLink {
			text += entity.URL + " "
		}
	}
	if message.From.ID == 777000 {
		return sources.FindSourceURL(text)
	}
	for _, entity := range message.CaptionEntities {
		if entity.Type == telego.EntityTypeTextLink {
			text += entity.URL + " "
		}
	}
	return sources.FindSourceURL(text)
}

func UpdateLinkPreview(ctx context.Context, targetMessage *telego.Message, artwork *types.Artwork, bot *telego.Bot, pictureIndex uint, photoParams *telego.SendPhotoParams) error {
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
		replyMarkup = telegram.GetPostedPictureReplyMarkup(artwork.Pictures[pictureIndex])
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
