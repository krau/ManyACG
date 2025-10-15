package utils

import (
	"context"
	"fmt"
	"html"
	"image"

	"github.com/gabriel-vasile/mimetype"
	"github.com/krau/ManyACG/internal/common/httpclient"
	"github.com/krau/ManyACG/internal/interface/telegram/metautil"
	"github.com/krau/ManyACG/internal/model/command"
	"github.com/krau/ManyACG/internal/model/entity"
	"github.com/krau/ManyACG/internal/model/query"
	"github.com/krau/ManyACG/internal/pkg/imgtool"
	"github.com/krau/ManyACG/internal/service"
	"github.com/krau/ManyACG/internal/shared"
	"github.com/krau/ManyACG/pkg/log"
	"github.com/krau/ManyACG/pkg/objectuuid"
	"github.com/krau/ManyACG/pkg/strutil"
	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegoutil"
	"github.com/samber/oops"
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
	showProgress := (fromChatID.ID != 0 || fromChatID.Username != "")
	useEdit := (messageID != 0)

	editReplyMarkupText := func(text string) {
		if !showProgress || !useEdit {
			return
		}
		_, err := bot.EditMessageReplyMarkup(ctx, telegoutil.EditMessageReplyMarkup(
			fromChatID,
			messageID,
			telegoutil.InlineKeyboard([]telego.InlineKeyboardButton{
				telegoutil.InlineKeyboardButton(text).WithCallbackData("noop"),
			}),
		))
		if err != nil {
			log.Warn("failed to edit reply markup", "err", err)
		}
	}
	// replyWaitMsg 回复 messageID 的消息
	replyWaitMsg := func(text string) {
		if !showProgress {
			return
		}
		if useEdit {
			_, err := bot.SendMessage(ctx, telegoutil.Message(fromChatID, text).WithReplyParameters(&telego.ReplyParameters{
				MessageID: messageID,
			}).WithParseMode(telego.ModeHTML))
			if err != nil {
				log.Warn("failed to send reply wait message", "err", err)
			}
			return
		}
		_, err := bot.SendMessage(ctx, telegoutil.Message(fromChatID, text).WithParseMode(telego.ModeHTML))
		if err != nil {
			log.Warn("failed to send reply wait message", "err", err)
		}
	}

	editReplyMarkupText("正在存储图片...")

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
			if info != nil {
				artwork.Pictures[i].StorageInfo = *info
			}
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
				origZip := ugoira.MetaData.OriginalZip
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
				ugoira.OriginalStorage = *info
				return nil
			}()
			if err != nil {
				return err
			}
		}
	}

	editReplyMarkupText("正在发布到频道...")

	msgs, err := SendArtworkMediaGroup(ctx, bot, serv, toChatID, artwork)
	if err != nil {
		return oops.Wrapf(err, "failed to send artwork media group")
	}
	if len(msgs) == 0 {
		return oops.New("no messages sent")
	}
	// 更新 cached artwork 的 TelegramInfo
	for _, msg := range msgs {
		tginfo := shared.TelegramInfo{
			MessageID:    msg.Message.MessageID,
			MediaGroupID: msg.Message.MediaGroupID,
			PhotoFileID:  msg.FileID,
		}
		if msg.UgoiraIndex >= 0 {
			artwork.UgoiraMetas[msg.UgoiraIndex].TelegramInfo = tginfo
		} else if msg.PictureIndex >= 0 {
			artwork.Pictures[msg.PictureIndex].TelegramInfo = tginfo
		} else {
			log.Warn("message has neither picture index nor ugoira index", "message_id", msg.Message.MessageID)
		}
	}
	if err := serv.UpdateCachedArtwork(ctx, artwork); err != nil {
		return oops.Wrapf(err, "failed to update cached artwork after sending")
	}
	// 创建 artwork
	awId, err := objectuuid.FromObjectIDHex(artwork.ID)
	if err != nil {
		awId = objectuuid.New()
	}
	ent, err := serv.CreateArtwork(ctx, &command.ArtworkCreation{
		ID:          awId,
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
					Data:            ugoira.MetaData,
					OriginalStorage: ugoira.OriginalStorage,
					TelegramInfo:    ugoira.TelegramInfo,
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

	if serv.ShouldTagNewArtwork() {
		log.Info("predicting artwork tags", "id", ent.ID, "title", ent.Title)
		editReplyMarkupText("已发布到频道, 正在推理作品标签...")
		if err := serv.PredictAndUpdateArtworkTags(ctx, ent.ID); err != nil {
			log.Error("failed to predict and update artwork tags after create artwork", "id", ent.ID, "err", err)
			editReplyMarkupText("推理作品标签失败, 作品已发布")
		}
		caption := ArtworkHTMLCaption(ent)
		bot.EditMessageCaption(ctx, telegoutil.
			EditMessageCaption(toChatID,
				ent.Pictures[0].TelegramInfo.Data().MessageID,
				caption).
			WithParseMode(telego.ModeHTML))
	}
	editReplyMarkupText("已发布到频道, 正在检测重复图片...")
	meta := metautil.FromContext(ctx)
	for i, pic := range ent.Pictures {
		similars, err := serv.QueryPicturesByPhash(ctx, query.PicturesPhash{Input: pic.Phash, Distance: 10})
		if err != nil {
			log.Error("failed to query pictures by phash", "phash", pic.Phash, "err", err)
			editReplyMarkupText(fmt.Sprintf("检测第%d张图片重复失败, 作品已发布", i+1))
			continue
		}
		if len(similars) == 0 {
			continue
		}
		sims := make([]*entity.Picture, 0, len(similars))
		for _, sim := range similars {
			if sim.ArtworkID == ent.ID {
				continue
			}
			sims = append(sims, sim)
		}
		if len(sims) == 0 {
			continue
		}
		log.Info("found similar pictures", "artwork_id", ent.ID, "picture_id", pic.ID, "count", len(sims), "similars", func() []string {
			ids := make([]string, len(sims))
			for i, s := range sims {
				ids[i] = s.ID.String()
			}
			return ids
		}())
		text := fmt.Sprintf("检测到 %d 张与作品 <a href='%s'>%s 第 %d 张图片</a>相似的图片", len(sims), func() string {
			if meta.ChannelAvailable() {
				return meta.ChannelMessageURL(pic.TelegramInfo.Data().MessageID)
			}
			return ent.SourceURL
		}(), html.EscapeString(ent.Title), pic.OrderIndex)
		for j, sim := range sims {
			text += fmt.Sprintf("\n\n%d - <a href='%s'>%s_%d</a>", j+1, func() string {
				if meta.ChannelAvailable() {
					return meta.ChannelMessageURL(sim.TelegramInfo.Data().MessageID)
				}
				return sim.Artwork.SourceURL
			}(), html.EscapeString(sim.Artwork.Title), sim.OrderIndex)
		}
		replyWaitMsg(text)
	}
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
