package utils

import (
	"context"
	"fmt"
	"image"

	"github.com/gabriel-vasile/mimetype"
	"github.com/krau/ManyACG/internal/common/httpclient"
	"github.com/krau/ManyACG/internal/model/command"
	"github.com/krau/ManyACG/internal/model/entity"
	"github.com/krau/ManyACG/internal/pkg/imgtool"
	"github.com/krau/ManyACG/internal/service"
	"github.com/krau/ManyACG/internal/shared"
	"github.com/krau/ManyACG/pkg/log"
	"github.com/krau/ManyACG/pkg/strutil"
	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegoutil"
	"github.com/samber/oops"
	"gorm.io/datatypes"
)

func doPostAndCreateArtwork(
	ctx context.Context,
	bot *telego.Bot,
	serv *service.Service,
	artwork *entity.CachedArtworkData,
	fromChatID telego.ChatID,
	toChatID telego.ChatID,
	messageID int,
) error {
	awInDb, err := serv.GetArtworkByURL(ctx, artwork.SourceURL)
	if err == nil {
		return oops.Errorf("artwork already exists in db: %s", awInDb.SourceURL)
	}
	if serv.CheckDeletedByURL(ctx, artwork.SourceURL) {
		return oops.Errorf("artwork is marked as deleted: %s", artwork.SourceURL)
	}
	showProgress := (fromChatID.ID != 0 || fromChatID.Username != "") && messageID != 0
	if showProgress {
		bot.EditMessageReplyMarkup(ctx, telegoutil.EditMessageReplyMarkup(
			fromChatID,
			messageID,
			telegoutil.InlineKeyboard([]telego.InlineKeyboardButton{
				telegoutil.InlineKeyboardButton("正在存储图片...").WithCallbackData("noop"),
			}),
		))
	}

	for i, pic := range artwork.Pictures {
		// 下载并存储图片, 同时计算 phash, thumbhash, width, height
		err = func() error {
			file, err := httpclient.DownloadWithCache(ctx, pic.Original, nil)
			if err != nil {
				return oops.Wrapf(err, "failed to download picture %d", i)
			}
			defer file.Close()
			img, _, err := image.Decode(file)
			if err != nil {
				return oops.Wrapf(err, "failed to decode picture %d", i)
			}
			if pic.Phash == "" {
				phash, err := imgtool.GetImagePhash(img)
				if err != nil {
					return oops.Wrapf(err, "failed to get phash of picture %d", i)
				}
				pic.Phash = phash
			}
			if pic.Width == 0 || pic.Height == 0 {
				w, h, err := imgtool.GetSize(img)
				if err != nil {
					return oops.Wrapf(err, "failed to get size of picture %d", i)
				}
				pic.Width = uint(w)
				pic.Height = uint(h)
			}
			if pic.ThumbHash == "" {
				thumbHash, err := imgtool.GetImageThumbHash(img)
				if err != nil {
					return oops.Wrapf(err, "failed to get thumb hash of picture %d", i)
				}
				pic.ThumbHash = thumbHash
			}
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
			info, err := serv.StorageSaveAllSize(ctx, file.Name(), fmt.Sprintf("/%s/%s", artwork.SourceType, artwork.Artist.UID), filename)
			if err != nil {
				return oops.Wrapf(err, "failed to save picture %d", i)
			}
			artwork.Pictures[i].StorageInfo = *info
			return nil
		}()
		if err != nil {
			return err
		}
	}
	isUgoira := len(artwork.UgoiraMetas) > 0
	if isUgoira {
		// 处理 ugoira 的 original
		for _, ugoira := range artwork.UgoiraMetas {
			err := func() error {
				origZip := ugoira.UgoiraMetaData.Data().OriginalZip
				file, err := httpclient.DownloadWithCache(ctx, origZip, nil)
				if err != nil {
					return oops.Wrapf(err, "failed to download ugoira original zip")
				}
				defer file.Close()
				filename := fmt.Sprintf("%s.zip", strutil.MD5Hash(origZip))
				info, err := serv.StorageSaveOriginal(ctx, file, fmt.Sprintf("/%s/%s/ugoira", artwork.SourceType, artwork.Artist.UID), filename)
				if err != nil {
					return oops.Wrapf(err, "failed to save ugoira original zip")
				}
				ugoira.OriginalStorage = datatypes.NewJSONType(*info)
				return nil
			}()
			if err != nil {
				return err
			}
		}
	}

	if showProgress {
		bot.EditMessageReplyMarkup(ctx, telegoutil.EditMessageReplyMarkup(
			fromChatID,
			messageID,
			telegoutil.InlineKeyboard([]telego.InlineKeyboardButton{
				telegoutil.InlineKeyboardButton("正在发布到频道...").WithCallbackData("noop"),
			}),
		))
	}

	msgs, err := SendArtworkPhotoMediaGroup(ctx, bot, toChatID, artwork)
	if err != nil {
		return oops.Wrapf(err, "failed to send artwork media group")
	}
	if len(msgs) == 0 {
		return oops.New("no messages sent")
	}
	// 更新 cached artwork 的 TelegramInfo
	for i := range artwork.Pictures {
		tginfo := shared.TelegramInfo{
			MessageID:    msgs[i].MessageID,
			MediaGroupID: msgs[i].MediaGroupID,
		}
		if photoSize := msgs[i].Photo; len(photoSize) > 0 {
			tginfo.PhotoFileID = photoSize[len(photoSize)-1].FileID
		}
		artwork.Pictures[i].TelegramInfo = tginfo
	}
	if err := serv.UpdateCachedArtwork(ctx, artwork); err != nil {
		return oops.Wrapf(err, "failed to update cached artwork after sending")
	}
	// 创建 artwork
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
		UgoiraMetas: func() []*command.ArtworkUgoiraCreation {
			if !isUgoira {
				return nil
			}
			ugos := make([]*command.ArtworkUgoiraCreation, len(artwork.UgoiraMetas))
			for i, ugoira := range artwork.UgoiraMetas {
				ugos[i] = &command.ArtworkUgoiraCreation{
					Index:           ugoira.OrderIndex,
					Data:            ugoira.UgoiraMetaData.Data(),
					OriginalStorage: ugoira.OriginalStorage.Data(),
					TelegramInfo:    ugoira.TelegramInfo.Data(),
				}
			}
			return ugos
		}(),
		Pictures: func() []command.ArtworkPictureCreation {
			pics := make([]command.ArtworkPictureCreation, len(artwork.Pictures))
			for i, pic := range artwork.Pictures {
				pics[i] = command.ArtworkPictureCreation{
					Index:        pic.OrderIndex,
					Thumbnail:    pic.Thumbnail,
					Original:     pic.Original,
					Width:        pic.Width,
					Height:       pic.Height,
					Phash:        pic.Phash,
					ThumbHash:    pic.ThumbHash,
					TelegramInfo: pic.TelegramInfo,
					StorageInfo:  pic.StorageInfo,
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

type postArtworkJob struct {
	ctx        context.Context
	bot        *telego.Bot
	serv       *service.Service
	artwork    *entity.CachedArtworkData
	fromChatID telego.ChatID
	toChatID   telego.ChatID
	messageID  int
	done       chan error
}

var (
	postArtworkTaskQueue chan *postArtworkJob
)

func init() {
	const (
		workerCount = 3
		queueSize   = 17
	)
	postArtworkTaskQueue = make(chan *postArtworkJob, queueSize)
	for i := 0; i < workerCount; i++ {
		go artworkPoster(i)
	}
}

func artworkPoster(id int) {
	for j := range postArtworkTaskQueue {
		err := doPostAndCreateArtwork(j.ctx, j.bot, j.serv, j.artwork, j.fromChatID, j.toChatID, j.messageID)
		if j.done != nil {
			j.done <- err
		}
	}
}

func PostAndCreateArtwork(
	ctx context.Context,
	bot *telego.Bot,
	serv *service.Service,
	artwork *entity.CachedArtworkData,
	fromChatID telego.ChatID,
	toChatID telego.ChatID,
	messageID int,
) error {
	done := make(chan error, 1)
	postArtworkTaskQueue <- &postArtworkJob{
		ctx:        ctx,
		bot:        bot,
		serv:       serv,
		artwork:    artwork,
		fromChatID: fromChatID,
		toChatID:   toChatID,
		messageID:  messageID,
		done:       done,
	}
	return <-done
}
