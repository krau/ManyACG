package utils

import (
	"context"

	"github.com/krau/ManyACG/internal/common/httpclient"
	"github.com/krau/ManyACG/internal/interface/telegram/metautil"
	"github.com/krau/ManyACG/internal/pkg/mediatool"
	"github.com/krau/ManyACG/internal/service"
	"github.com/krau/ManyACG/internal/shared"
	"github.com/krau/ManyACG/pkg/ioutil"
	"github.com/krau/ManyACG/pkg/osutil"
	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegoutil"
	"github.com/samber/oops"
)

type MediaResultType uint

const (
	MediaResultTypePhoto MediaResultType = iota + 1
	MediaResultTypeUgoira
	MediaResultTypeVideo // although ugoira is also be sent as video, we need to distinguish them for file ID update purposes
)

type MediaGroupResultMessage struct {
	Message telego.Message
	FileID  string // Telegram file ID
	Type    MediaResultType
	Index   int
}

func SendArtworkMediaGroup(
	ctx context.Context,
	bot *telego.Bot,
	serv *service.Service,
	meta *metautil.MetaData,
	chatID telego.ChatID,
	artwork shared.ArtworkLike) ([]MediaGroupResultMessage, error) {

	go bot.SendChatAction(ctx, telegoutil.ChatAction(chatID, telego.ChatActionUploadPhoto))

	results := make([]MediaGroupResultMessage, 0)
	photoMsgs, err := SendArtworkPhotoMediaGroup(ctx, bot, serv, meta, chatID, artwork)
	if err != nil {
		return nil, oops.Wrapf(err, "failed to send artwork photo media group")
	}
	// collect photo file ids
	for i, msg := range photoMsgs {
		if len(msg.Photo) == 0 {
			continue
		}
		results = append(results, MediaGroupResultMessage{
			Message: msg,
			FileID:  msg.Photo[len(msg.Photo)-1].FileID, // get the highest resolution photo
			Index:   i,
			Type:    MediaResultTypePhoto,
		})
	}
	if mediatool.FFmpegAvailable() && len(artwork.GetUgoiraMetas()) > 0 {
		sendOption := &SendOption{}
		if len(results) > 0 {
			sendOption.ReplyTo = results[0].Message.MessageID
		}
		ugoiraMsgs, err := SendArtworkUgoiraMediaGroup(ctx, bot, serv, meta, chatID, artwork, sendOption)
		if err != nil {
			return nil, oops.Wrapf(err, "failed to send artwork ugoira media group")
		}
		// collect ugoira file ids
		for i, msg := range ugoiraMsgs {
			if msg.Video == nil {
				continue
			}
			results = append(results, MediaGroupResultMessage{
				Message: msg,
				FileID:  msg.Video.FileID,
				Type:    MediaResultTypeUgoira,
				Index:   i,
			})
		}
	}
	if len(artwork.GetVideos()) > 0 {
		sendOption := &SendOption{}
		if len(results) > 0 {
			sendOption.ReplyTo = results[0].Message.MessageID
		}
		videoMsgs, err := SendArtworkVideoMediaGroup(ctx, bot, serv, meta, chatID, artwork, sendOption)
		if err != nil {
			return nil, oops.Wrapf(err, "failed to send artwork video media group")
		}
		// collect video file ids
		for i, msg := range videoMsgs {
			if msg.Video == nil {
				continue
			}
			results = append(results, MediaGroupResultMessage{
				Message: msg,
				FileID:  msg.Video.FileID,
				Type:    MediaResultTypeVideo,
				Index:   i,
			})
		}
	}

	return results, nil
}

func SendArtworkPhotoMediaGroup(
	ctx context.Context,
	bot *telego.Bot,
	serv *service.Service,
	meta *metautil.MetaData,
	chatID telego.ChatID,
	artwork shared.ArtworkLike) ([]telego.Message, error) {

	pics := artwork.GetPictures()
	caption := ArtworkHTMLCaption(artwork)

	if len(pics) <= 10 {
		inputs, err := ArtworkInputMediaPhotos(ctx, serv, meta, artwork, caption, 0, len(pics))
		if err != nil {
			return nil, oops.Wrapf(err, "failed to create input media photos")
		}
		defer inputs.Close()
		// Send the media group
		return bot.SendMediaGroup(ctx, telegoutil.MediaGroup(
			chatID,
			inputs.Value...,
		))
	}
	messages := make([]telego.Message, len(pics))
	for i := 0; i < len(pics); i += 10 {
		end := i + 10
		if end > len(pics) {
			end = len(pics)
		}
		inputs, err := ArtworkInputMediaPhotos(ctx, serv, meta, artwork, caption, i, end)
		if err != nil {
			return nil, oops.Wrapf(err, "failed to create input media photos")
		}
		defer inputs.Close()
		mediaGroup := telegoutil.MediaGroup(chatID, inputs.Value...)
		if i > 0 {
			mediaGroup = mediaGroup.WithReplyParameters(&telego.ReplyParameters{
				ChatID:    chatID,
				MessageID: messages[i-1].MessageID,
			})
		}
		msgs, err := bot.SendMediaGroup(ctx, mediaGroup)
		if err != nil {
			return nil, oops.Wrapf(err, "failed to send media group")
		}
		copy(messages[i:], msgs)
	}
	return messages, nil
}

func ArtworkInputMediaPhotos(
	ctx context.Context,
	serv *service.Service,
	meta *metautil.MetaData,
	artwork shared.ArtworkLike,
	caption string,
	start, end int) (*ioutil.Closer[[]telego.InputMedia], error) {
	inputMediaPhotos := make([]telego.InputMedia, end-start)
	awPics := artwork.GetPictures()
	if start < 0 || end > len(awPics) || start >= end {
		return nil, oops.Errorf("invalid start or end index: %d, %d, len=%d", start, end, len(awPics))
	}

	closers := make([]func() error, 0, end-start)

	for i := start; i < end; i++ {
		picture := awPics[i]
		err := func() error {
			var photo *telego.InputMediaPhoto
			if id := picture.GetTelegramInfo().PhotoFileID(meta.BotID()); id != "" {
				photo = telegoutil.MediaPhoto(telegoutil.FileFromID(id))
			} else {
				if picture.GetStorageInfo() != shared.ZeroStorageInfo && picture.GetStorageInfo().Original != nil {
					// download from storage
					file, err := serv.StorageGetFile(ctx, *picture.GetStorageInfo().Original)
					if err != nil {
						return oops.Wrapf(err, "failed to get file from storage")
					}
					defer file.Close()
					compressed, err := mediatool.CompressForTelegramFromFile(file.Name())
					if err != nil {
						return oops.Wrapf(err, "failed to compress image")
					}
					photo = telegoutil.MediaPhoto(telegoutil.File(compressed))
					closers = append(closers, func() error { return compressed.Close() })
				} else {
					file, err := httpclient.DownloadWithCache(ctx, picture.GetOriginal(), nil)
					if err != nil {
						return oops.Wrapf(err, "failed to download file: %s", picture.GetOriginal())
					}
					defer file.Close()
					compressed, err := mediatool.CompressForTelegramFromFile(file.Name())
					if err != nil {
						return oops.Wrapf(err, "failed to compress image")
					}
					photo = telegoutil.MediaPhoto(telegoutil.File(compressed))
					closers = append(closers, func() error { return compressed.Close() })
				}
			}
			if photo == nil {
				return oops.New("failed to create input media photo")
			}
			if i == 0 {
				photo = photo.WithCaption(caption).WithParseMode(telego.ModeHTML)
			}
			if artwork.GetR18() {
				photo = photo.WithHasSpoiler()
			}
			inputMediaPhotos[i-start] = photo
			return nil
		}()
		if err != nil {
			var closeErrs []error
			for _, closer := range closers {
				if err := closer(); err != nil {
					closeErrs = append(closeErrs, err)
				}
			}
			return nil, oops.Wrapf(err, "failed to create input media photo, close errs: %v", oops.Join(closeErrs...))
		}
	}
	return &ioutil.Closer[[]telego.InputMedia]{
		Value: inputMediaPhotos,
		CloseFunc: func() error {
			var errs []error
			for _, closer := range closers {
				if err := closer(); err != nil {
					errs = append(errs, err)
				}
			}
			return oops.Join(errs...)
		},
	}, nil
}

type SendOption struct {
	ReplyTo int
}

func SendArtworkUgoiraMediaGroup(
	ctx context.Context,
	bot *telego.Bot,
	serv *service.Service,
	meta *metautil.MetaData,
	chatID telego.ChatID,
	artwork shared.ArtworkLike,
	opt *SendOption,
) ([]telego.Message, error) {

	ugoiras := artwork.GetUgoiraMetas()
	caption := ArtworkHTMLCaption(artwork)
	if len(ugoiras) <= 10 {
		inputs, err := ArtworkInputMediaVideos(ctx, serv, meta, artwork, caption, 0, len(ugoiras))
		if err != nil {
			return nil, oops.Wrapf(err, "failed to create input media videos")
		}
		defer inputs.Close()
		mediaGroup := telegoutil.MediaGroup(
			chatID,
			inputs.Value...,
		)
		if opt != nil && opt.ReplyTo != 0 {
			mediaGroup = mediaGroup.WithReplyParameters(&telego.ReplyParameters{
				ChatID:    chatID,
				MessageID: opt.ReplyTo,
			})
		}
		return bot.SendMediaGroup(ctx, mediaGroup)
	}

	messages := make([]telego.Message, len(ugoiras))
	for i := 0; i < len(ugoiras); i += 10 {
		end := i + 10
		if end > len(ugoiras) {
			end = len(ugoiras)
		}
		inputs, err := ArtworkInputMediaVideos(ctx, serv, meta, artwork, caption, i, end)
		if err != nil {
			return nil, oops.Wrapf(err, "failed to create input media videos")
		}
		defer inputs.Close()
		mediaGroup := telegoutil.MediaGroup(chatID, inputs.Value...)
		if opt != nil && opt.ReplyTo != 0 && i == 0 {
			mediaGroup = mediaGroup.WithReplyParameters(&telego.ReplyParameters{
				ChatID:    chatID,
				MessageID: opt.ReplyTo,
			})
		}
		if i > 0 {
			mediaGroup = mediaGroup.WithReplyParameters(&telego.ReplyParameters{
				ChatID:    chatID,
				MessageID: messages[i-1].MessageID,
			})
		}
		msgs, err := bot.SendMediaGroup(ctx, mediaGroup)
		if err != nil {
			return nil, oops.Wrapf(err, "failed to send media group")
		}
		copy(messages[i:], msgs)
	}
	return messages, nil
}

func ArtworkInputMediaVideos(ctx context.Context,
	serv *service.Service,
	meta *metautil.MetaData,
	artwork shared.ArtworkLike,
	caption string,
	start, end int) (*ioutil.Closer[[]telego.InputMedia], error) {

	if start < 0 || start >= end {
		return nil, oops.Errorf("invalid start or end index: %d, %d", start, end)
	}
	inputMedias := make([]telego.InputMedia, end-start)
	closers := make([]func() error, 0, end-start)

	if ugoiras := artwork.GetUgoiraMetas(); len(ugoiras) > 0 {
		if end > len(ugoiras) {
			return nil, oops.Errorf("invalid start or end index: %d, %d, len=%d", start, end, len(ugoiras))
		}

		for i := start; i < end; i++ {
			ugoira := ugoiras[i]
			err := func() error {
				var video *telego.InputMediaVideo
				if id := ugoira.GetTelegramInfo().VideoFileID(meta.BotID()); id != "" {
					video = telegoutil.MediaVideo(telegoutil.FileFromID(id))
				} else {
					storDetail := ugoira.GetOriginalStorage()
					if storDetail != shared.ZeroStorageDetail {
						// download from storage
						file, err := serv.StorageGetFile(ctx, storDetail)
						if err != nil {
							return oops.Wrapf(err, "failed to get file from storage")
						}
						defer file.Close()
						videoPath, err := mediatool.UgoiraZipToMp4(file.Name(), ugoira.GetUgoiraMetaData().Frames, file.Name()+".mp4")
						if err != nil {
							return oops.Wrapf(err, "failed to compress image")
						}
						videoFile, err := osutil.OpenTemp(videoPath)
						if err != nil {
							return oops.Wrapf(err, "failed to open mp4 file")
						}
						video = telegoutil.MediaVideo(telegoutil.File(videoFile))
						closers = append(closers, func() error { return videoFile.Close() })
					} else {
						file, err := httpclient.DownloadWithCache(ctx, ugoira.GetUgoiraMetaData().OriginalZip, nil)
						if err != nil {
							return oops.Wrapf(err, "failed to download file: %s", ugoira.GetUgoiraMetaData().OriginalZip)
						}
						defer file.Close()
						videoPath, err := mediatool.UgoiraZipToMp4(file.Name(), ugoira.GetUgoiraMetaData().Frames, file.Name()+".mp4")
						if err != nil {
							return oops.Wrapf(err, "failed to compress image")
						}
						videoFile, err := osutil.OpenTemp(videoPath)
						if err != nil {
							return oops.Wrapf(err, "failed to open mp4 file")
						}
						video = telegoutil.MediaVideo(telegoutil.File(videoFile))
						closers = append(closers, func() error { return videoFile.Close() })
					}
				}
				if video == nil {
					return oops.New("failed to create input media video")
				}
				if i == 0 {
					video = video.WithCaption(caption).WithParseMode(telego.ModeHTML)
				}
				if artwork.GetR18() {
					video = video.WithHasSpoiler()
				}
				inputMedias[i-start] = video
				return nil
			}()
			if err != nil {
				var closeErrs []error
				for _, close := range closers {
					if err := close(); err != nil {
						closeErrs = append(closeErrs, err)
					}
				}
				return nil, oops.Wrapf(err, "failed to create input media photo, close errs: %v", oops.Join(closeErrs...))
			}
		}
		return &ioutil.Closer[[]telego.InputMedia]{
			Value: inputMedias,
			CloseFunc: func() error {
				var errs []error
				for _, close := range closers {
					if err := close(); err != nil {
						errs = append(errs, err)
					}
				}
				return oops.Join(errs...)
			},
		}, nil
	}

	if videos := artwork.GetVideos(); len(videos) > 0 {
		if end > len(videos) {
			return nil, oops.Errorf("invalid start or end index: %d, %d, len=%d", start, end, len(videos))
		}

		for i := start; i < end; i++ {
			video := videos[i]
			err := func() error {
				var mediaVideo *telego.InputMediaVideo
				if id := video.GetTelegramInfo().VideoFileID(meta.BotID()); id != "" {
					mediaVideo = telegoutil.MediaVideo(telegoutil.FileFromID(id))
				} else {
					file, err := httpclient.DownloadWithCache(ctx, video.GetURL(), nil)
					if err != nil {
						return oops.Wrapf(err, "failed to download file: %s", video.GetURL())
					}
					defer file.Close()
					videoFile, err := osutil.OpenTemp(file.Name())
					if err != nil {
						return oops.Wrapf(err, "failed to open video file")
					}
					mediaVideo = telegoutil.MediaVideo(telegoutil.File(videoFile))
					closers = append(closers, func() error { return videoFile.Close() })
				}
				if mediaVideo == nil {
					return oops.New("failed to create input media video")
				}
				if i == 0 {
					mediaVideo = mediaVideo.WithCaption(caption).WithParseMode(telego.ModeHTML)
				}
				if artwork.GetR18() {
					mediaVideo = mediaVideo.WithHasSpoiler()
				}
				inputMedias[i-start] = mediaVideo
				return nil
			}()
			if err != nil {
				var closeErrs []error
				for _, close := range closers {
					if err := close(); err != nil {
						closeErrs = append(closeErrs, err)
					}
				}
				return nil, oops.Wrapf(err, "failed to create input media video, close errs: %v", oops.Join(closeErrs...))
			}
		}
		return &ioutil.Closer[[]telego.InputMedia]{
			Value: inputMedias,
			CloseFunc: func() error {
				var errs []error
				for _, close := range closers {
					if err := close(); err != nil {
						errs = append(errs, err)
					}
				}
				return oops.Join(errs...)
			},
		}, nil
	}
	return nil, oops.New("no ugoira or video found")
}

// [TODO] this seems duplicated with SendArtworkUgoiraMediaGroup
func SendArtworkVideoMediaGroup(
	ctx context.Context,
	bot *telego.Bot,
	serv *service.Service,
	meta *metautil.MetaData,
	chatID telego.ChatID,
	artwork shared.ArtworkLike,
	opt *SendOption,
) ([]telego.Message, error) {
	videos := artwork.GetVideos()
	caption := ArtworkHTMLCaption(artwork)

	if len(videos) <= 10 {
		inputs, err := ArtworkInputMediaVideos(ctx, serv, meta, artwork, caption, 0, len(videos))
		if err != nil {
			return nil, oops.Wrapf(err, "failed to create input media videos")
		}
		defer inputs.Close()
		mediaGroup := telegoutil.MediaGroup(
			chatID,
			inputs.Value...,
		)
		if opt != nil && opt.ReplyTo != 0 {
			mediaGroup = mediaGroup.WithReplyParameters(&telego.ReplyParameters{
				ChatID:    chatID,
				MessageID: opt.ReplyTo,
			})
		}
		return bot.SendMediaGroup(ctx, mediaGroup)
	}
	messages := make([]telego.Message, len(videos))
	for i := 0; i < len(videos); i += 10 {
		end := i + 10
		if end > len(videos) {
			end = len(videos)
		}
		inputs, err := ArtworkInputMediaVideos(ctx, serv, meta, artwork, caption, i, end)
		if err != nil {
			return nil, oops.Wrapf(err, "failed to create input media videos")
		}
		defer inputs.Close()
		mediaGroup := telegoutil.MediaGroup(chatID, inputs.Value...)
		if opt != nil && opt.ReplyTo != 0 && i == 0 {
			mediaGroup = mediaGroup.WithReplyParameters(&telego.ReplyParameters{
				ChatID:    chatID,
				MessageID: opt.ReplyTo,
			})
		}
		if i > 0 {
			mediaGroup = mediaGroup.WithReplyParameters(&telego.ReplyParameters{
				ChatID:    chatID,
				MessageID: messages[i-1].MessageID,
			})
		}
		msgs, err := bot.SendMediaGroup(ctx, mediaGroup)
		if err != nil {
			return nil, oops.Wrapf(err, "failed to send media group")
		}
		copy(messages[i:], msgs)
	}
	return messages, nil
}
