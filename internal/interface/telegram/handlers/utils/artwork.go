package utils

import (
	"bytes"
	"context"
	"fmt"
	"os"

	"github.com/gabriel-vasile/mimetype"
	"github.com/google/uuid"
	"github.com/krau/ManyACG/internal/common/httpclient"
	"github.com/krau/ManyACG/internal/interface/telegram/metautil"
	"github.com/krau/ManyACG/internal/model/command"
	"github.com/krau/ManyACG/internal/model/entity"
	"github.com/krau/ManyACG/internal/pkg/imgtool"
	"github.com/krau/ManyACG/internal/shared"
	"github.com/krau/ManyACG/pkg/log"
	"github.com/krau/ManyACG/pkg/strutil"
	"github.com/krau/ManyACG/service"
	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegohandler"
	"github.com/mymmrac/telego/telegoutil"
	"github.com/samber/oops"
)

func PostAndCreateArtwork(
	ctx *telegohandler.Context,
	serv *service.Service,
	artwork *entity.CachedArtworkData,
	fromChatID int64, messageID int,
) error {
	awInDb, err := serv.GetArtworkByURL(ctx, artwork.SourceURL)
	if err == nil {
		return oops.Errorf("artwork already exists in db: %s", awInDb.SourceURL)
	}
	if serv.CheckDeletedByURL(ctx, artwork.SourceURL) {
		return oops.Errorf("artwork is marked as deleted: %s", artwork.SourceURL)
	}
	showProgress := fromChatID != 0 && messageID != 0
	if showProgress {
		// // clear previous inline buttons
		// defer ctx.Bot().EditMessageReplyMarkup(ctx, &telego.EditMessageReplyMarkupParams{
		// 	ChatID:      telegoutil.ID(fromChatID),
		// 	MessageID:   messageID,
		// 	ReplyMarkup: nil,
		// })
		ctx.Bot().EditMessageReplyMarkup(ctx, telegoutil.EditMessageReplyMarkup(
			telegoutil.ID(fromChatID),
			messageID,
			telegoutil.InlineKeyboard([]telego.InlineKeyboardButton{
				telegoutil.InlineKeyboardButton("正在存储图片...").WithCallbackData("noop"),
			}),
		))
	}

	for i, pic := range artwork.Pictures {
		file, clean, err := httpclient.DownloadWithCache(ctx, pic.Original, nil)
		if err != nil {
			return oops.Wrapf(err, "failed to download picture %d", i)
		}
		defer file.Close()
		defer clean()
		var ext string
		ext, err = strutil.GetFileExtFromURL(pic.Original)
		if err != nil {
			mtype, err := mimetype.DetectFile(file.Name())
			if err != nil {
				return oops.Wrapf(err, "failed to detect mime type for picture %d", i)
			}
			ext = mtype.Extension()
		}
		filename := fmt.Sprintf("%s%s", strutil.MD5Hash(pic.Original), ext)
		info, err := serv.StorageSaveAllSize(ctx, file, fmt.Sprint("/%s/%s", artwork.SourceType, artwork.Artist.UID), filename)
		if err != nil {
			return oops.Wrapf(err, "failed to save picture %d", i)
		}
		artwork.Pictures[i].StorageInfo = *info
	}
	if showProgress {
		ctx.Bot().EditMessageReplyMarkup(ctx, telegoutil.EditMessageReplyMarkup(
			telegoutil.ID(fromChatID),
			messageID,
			telegoutil.InlineKeyboard([]telego.InlineKeyboardButton{
				telegoutil.InlineKeyboardButton("正在发布到频道...").WithCallbackData("noop"),
			}),
		))
	}
	msgs, err := SendArtworkMediaGroup(ctx, telegoutil.ID(fromChatID), artwork)
	if err != nil {
		return oops.Wrapf(err, "failed to send artwork media group")
	}
	if len(msgs) == 0 {
		return oops.New("no messages sent")
	}
	for i := range artwork.Pictures {
		tginfo := shared.TelegramInfo{
			MessageID:    msgs[i].MessageID,
			MediaGroupID: msgs[i].MediaGroupID,
		}
		if photo := msgs[i].Photo; len(photo) > 0 {
			tginfo.PhotoFileID = photo[len(photo)-1].FileID
		}
		if doc := msgs[i].Document; doc != nil {
			tginfo.DocumentFileID = doc.FileID
		}
		artwork.Pictures[i].TelegramInfo = tginfo
	}
	ent, err := serv.CreateArtwork(ctx, &command.ArtworkCreation{
		Title:       artwork.Title,
		Description: artwork.Description,
		R18:         artwork.R18,
		SourceType:  artwork.SourceType,
		Artist: command.ArtworkArtistCreation{
			Name:     artwork.Artist.Name,
			UID:      artwork.Artist.UID,
			Username: artwork.Artist.Username,
		},
		SourceURL: artwork.SourceURL,
		Tags:      artwork.Tags,
		Pictures: func() []command.ArtworkPictureCreation {
			pics := make([]command.ArtworkPictureCreation, len(artwork.Pictures))
			for i, pic := range artwork.Pictures {
				pics[i] = command.ArtworkPictureCreation{
					Index:        pic.Index,
					Thumbnail:    pic.Thumbnail,
					Original:     pic.Original,
					Width:        pic.Width,
					Height:       pic.Height,
					Phash:        pic.Phash,
					ThumbHash:    pic.ThumbHash,
					TelegramInfo: &pic.TelegramInfo,
					StorageInfo:  &pic.StorageInfo,
				}
			}
			return pics
		}(),
	})
	if err != nil {
		return oops.Wrapf(err, "failed to create artwork in db")
	}
	log.Info("created artwork", "id", ent.ID, "url", ent.SourceURL, "title", ent.Title, "pics", len(ent.Pictures))
	return nil
}

func SendArtworkMediaGroup(ctx *telegohandler.Context, chatID telego.ChatID, artwork entity.ArtworkLike) ([]telego.Message, error) {
	pics := artwork.GetPictures()
	if len(pics) <= 10 {
		inputs, err := ArtworkInputMediaPhotos(ctx, service.FromContext(ctx), artwork, ArtworkHTMLCaption(metautil.FromContext(ctx), artwork), 0, len(pics))
		if err != nil {
			return nil, oops.Wrapf(err, "failed to create input media photos")
		}
		// Send the media group
		return ctx.Bot().SendMediaGroup(ctx, telegoutil.MediaGroup(
			chatID,
			inputs...,
		))
	}
	caption := ArtworkHTMLCaption(metautil.FromContext(ctx), artwork)
	messages := make([]telego.Message, len(pics))
	for i := 0; i < len(pics); i += 10 {
		end := i + 10
		if end > len(pics) {
			end = len(pics)
		}
		inputs, err := ArtworkInputMediaPhotos(ctx, service.FromContext(ctx), artwork, caption, i, end)
		if err != nil {
			return nil, oops.Wrapf(err, "failed to create input media photos")
		}
		mediaGroup := telegoutil.MediaGroup(chatID, inputs...)
		if i > 0 {
			mediaGroup = mediaGroup.WithReplyParameters(&telego.ReplyParameters{
				ChatID:    chatID,
				MessageID: messages[i-1].MessageID,
			})
		}
		msgs, err := ctx.Bot().SendMediaGroup(ctx, mediaGroup)
		if err != nil {
			return nil, oops.Wrapf(err, "failed to send media group")
		}
		copy(messages[i:], msgs)
	}
	return messages, nil
}

func ArtworkInputMediaPhotos(ctx context.Context,
	serv *service.Service,
	artwork entity.ArtworkLike,
	caption string,
	start, end int) ([]telego.InputMedia, error) {
	inputMediaPhotos := make([]telego.InputMedia, end-start)
	for i := start; i < end; i++ {
		picture := artwork.GetPictures()[i]
		var photo *telego.InputMediaPhoto
		if id := picture.GetTelegramInfo().PhotoFileID; id != "" {
			photo = telegoutil.MediaPhoto(telegoutil.FileFromID(id))
		}
		if photo == nil {
			var fileBytes []byte
			var err error
			fileBytes, err = httpclient.GetReqCachedFile(picture.GetOriginal())
			if err != nil {
				if picture.GetStorageInfo() == shared.ZeroStorageInfo || picture.GetStorageInfo().Original == nil {
					file, clean, err := httpclient.DownloadWithCache(ctx, picture.GetOriginal(), nil)
					if err != nil {
						return nil, oops.Wrapf(err, "failed to download file: %s", picture.GetOriginal())
					}
					defer file.Close()
					defer clean()
					fileBytes, err = os.ReadFile(file.Name())
					if err != nil {
						return nil, oops.Wrapf(err, "failed to read file: %s", picture.GetOriginal())
					}
				} else {
					fileBytes, err = serv.Storage(picture.GetStorageInfo().Original.Type).GetFile(ctx, *picture.GetStorageInfo().Original)
					if err != nil {
						return nil, oops.Wrapf(err, "failed to get file: %s", picture.GetOriginal())
					}
				}
			}
			fileBytes, err = imgtool.CompressImageForTelegram(fileBytes)
			if err != nil {
				return nil, oops.Wrapf(err, "failed to compress image: %s", picture.GetOriginal())
			}
			photo = telegoutil.MediaPhoto(telegoutil.File(telegoutil.NameReader(bytes.NewReader(fileBytes), uuid.New().String())))
		}
		if i == 0 {
			photo = photo.WithCaption(caption).WithParseMode(telego.ModeHTML)
		}
		if artwork.GetR18() {
			photo = photo.WithHasSpoiler()
		}
		inputMediaPhotos[i-start] = photo
	}
	return inputMediaPhotos, nil
}

func GetPicturePreviewInputFile(ctx context.Context, picture entity.PictureLike) (telego.InputFile, error) {
	panic("not implemented")
}
