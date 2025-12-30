package utils

import (
	"context"
	"io"

	"github.com/krau/ManyACG/internal/common/httpclient"
	"github.com/krau/ManyACG/internal/interface/telegram/metautil"
	"github.com/krau/ManyACG/internal/pkg/mediatool"
	"github.com/krau/ManyACG/internal/service"
	"github.com/krau/ManyACG/internal/shared"
	"github.com/krau/ManyACG/pkg/ioutil"
	"github.com/krau/ManyACG/pkg/log"
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

// MediaItem represents a photo, ugoira or video for unified processing
type MediaItem struct {
	Picture      shared.PictureLike
	Ugoira       shared.UgoiraMetaLike
	Video        shared.VideoLike
	TelegramInfo shared.TelegramInfo
	Type         MediaResultType
	Index        int // original index in pictures, ugoiras or videos array
}

func SendArtworkMediaGroup(
	ctx context.Context,
	bot *telego.Bot,
	serv *service.Service,
	meta *metautil.MetaData,
	chatID telego.ChatID,
	artwork shared.ArtworkLike) ([]MediaGroupResultMessage, error) {

	go bot.SendChatAction(ctx, telegoutil.ChatAction(chatID, telego.ChatActionUploadPhoto))

	// https://core.telegram.org/bots/api#sendmediagroup
	// Photos and videos can be mixed in the same media group
	includeUgoira := mediatool.FFmpegAvailable() && len(artwork.GetUgoiraMetas()) > 0
	items := buildMediaItems(artwork, includeUgoira)
	if len(items) == 0 {
		return nil, oops.New("no media found in artwork")
	}

	caption := ArtworkHTMLCaption(artwork)
	results := make([]MediaGroupResultMessage, 0, len(items))

	if len(items) <= 10 {
		inputs, err := ArtworkInputMedias(ctx, serv, meta, artwork, caption, items, 0, len(items))
		if err != nil {
			return nil, oops.Wrapf(err, "failed to create input medias")
		}
		defer inputs.Close()
		msgs, err := bot.SendMediaGroup(ctx, telegoutil.MediaGroup(chatID, inputs.Value...))
		if err != nil {
			return nil, oops.Wrapf(err, "failed to send media group")
		}
		for i, msg := range msgs {
			result := MediaGroupResultMessage{
				Message: msg,
				Type:    items[i].Type,
				Index:   items[i].Index,
			}
			if items[i].Type == MediaResultTypePhoto && len(msg.Photo) > 0 {
				result.FileID = msg.Photo[len(msg.Photo)-1].FileID
			} else if msg.Video != nil {
				result.FileID = msg.Video.FileID
			}
			results = append(results, result)
		}
		return results, nil
	}

	// Send in batches of 10
	messages := make([]telego.Message, len(items))
	for i := 0; i < len(items); i += 10 {
		end := min(i+10, len(items))
		inputs, err := ArtworkInputMedias(ctx, serv, meta, artwork, caption, items, i, end)
		if err != nil {
			return nil, oops.Wrapf(err, "failed to create input medias")
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

	for i, msg := range messages {
		result := MediaGroupResultMessage{
			Message: msg,
			Type:    items[i].Type,
			Index:   items[i].Index,
		}
		if items[i].Type == MediaResultTypePhoto && len(msg.Photo) > 0 {
			result.FileID = msg.Photo[len(msg.Photo)-1].FileID
		} else if msg.Video != nil {
			result.FileID = msg.Video.FileID
		}
		results = append(results, result)
	}

	return results, nil
}

type SendOption struct {
	ReplyTo int
}

// buildMediaItems creates a combined list of MediaItems from pictures, ugoiras and videos
func buildMediaItems(artwork shared.ArtworkLike, includeUgoira bool) []MediaItem {
	items := make([]MediaItem, 0)

	// Add pictures first
	for i, pic := range artwork.GetPictures() {
		items = append(items, MediaItem{
			Type:         MediaResultTypePhoto,
			Index:        i,
			Picture:      pic,
			TelegramInfo: pic.GetTelegramInfo(),
		})
	}

	// Add ugoiras (converted to video)
	if includeUgoira {
		for i, ugoira := range artwork.GetUgoiraMetas() {
			items = append(items, MediaItem{
				Type:         MediaResultTypeUgoira,
				Index:        i,
				Ugoira:       ugoira,
				TelegramInfo: ugoira.GetTelegramInfo(),
			})
		}
	}

	// Add videos
	for i, video := range artwork.GetVideos() {
		items = append(items, MediaItem{
			Type:         MediaResultTypeVideo,
			Index:        i,
			Video:        video,
			TelegramInfo: video.GetTelegramInfo(),
		})
	}

	return items
}

// ArtworkInputMedias creates input medias from MediaItems (supports mixed photos and videos)
func ArtworkInputMedias(
	ctx context.Context,
	serv *service.Service,
	meta *metautil.MetaData,
	artwork shared.ArtworkLike,
	caption string,
	items []MediaItem,
	start, end int,
) (*ioutil.Closer[[]telego.InputMedia], error) {
	if start < 0 || end > len(items) || start >= end {
		return nil, oops.Errorf("invalid start or end index: %d, %d, len=%d", start, end, len(items))
	}

	inputMedias := make([]telego.InputMedia, end-start)
	closers := make([]func() error, 0, end-start)

	for i := start; i < end; i++ {
		item := items[i]
		err := func() error {
			var inputMedia telego.InputMedia

			switch item.Type {
			case MediaResultTypePhoto:
				picture := item.Picture
				var photo *telego.InputMediaPhoto
				if id := item.TelegramInfo.PhotoFileID(meta.BotID()); id != "" {
					photo = telegoutil.MediaPhoto(telegoutil.FileFromID(id))
				} else {
					if picture.GetStorageInfo() != shared.ZeroStorageInfo && picture.GetStorageInfo().Original != nil {
						file, err := serv.StorageGetFile(ctx, *picture.GetStorageInfo().Original)
						if err != nil {
							return oops.Wrapf(err, "failed to get file from storage")
						}
						defer file.Close()
						compressed, err := mediatool.CompressImgForTelegramFromFile(file.Name())
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
						compressed, err := mediatool.CompressImgForTelegramFromFile(file.Name())
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
				if i == start {
					photo = photo.WithCaption(caption).WithParseMode(telego.ModeHTML)
				}
				if artwork.GetR18() {
					photo = photo.WithHasSpoiler()
				}
				inputMedia = photo

			case MediaResultTypeUgoira:
				ugoira := item.Ugoira
				var video *telego.InputMediaVideo
				if id := item.TelegramInfo.VideoFileID(meta.BotID()); id != "" {
					video = telegoutil.MediaVideo(telegoutil.FileFromID(id))
				} else {
					storDetail := ugoira.GetOriginalStorage()
					if storDetail != shared.ZeroStorageDetail {
						file, err := serv.StorageGetFile(ctx, storDetail)
						if err != nil {
							return oops.Wrapf(err, "failed to get file from storage")
						}
						defer file.Close()
						videoPath, err := mediatool.UgoiraZipToMp4(file.Name(), ugoira.GetData().Frames, file.Name()+".mp4")
						if err != nil {
							return oops.Wrapf(err, "failed to convert ugoira to mp4")
						}
						videoFile, err := osutil.OpenTemp(videoPath)
						if err != nil {
							return oops.Wrapf(err, "failed to open mp4 file")
						}
						video = telegoutil.MediaVideo(telegoutil.File(videoFile))
						closers = append(closers, func() error { return videoFile.Close() })
					} else {
						file, err := httpclient.DownloadWithCache(ctx, ugoira.GetData().OriginalZip, nil)
						if err != nil {
							return oops.Wrapf(err, "failed to download file: %s", ugoira.GetData().OriginalZip)
						}
						defer file.Close()
						videoPath, err := mediatool.UgoiraZipToMp4(file.Name(), ugoira.GetData().Frames, file.Name()+".mp4")
						if err != nil {
							return oops.Wrapf(err, "failed to convert ugoira to mp4")
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
					return oops.New("failed to create input media video for ugoira")
				}
				if i == start {
					video = video.WithCaption(caption).WithParseMode(telego.ModeHTML)
				}
				if artwork.GetR18() {
					video = video.WithHasSpoiler()
				}
				inputMedia = video

			case MediaResultTypeVideo:
				v := item.Video
				var video *telego.InputMediaVideo
				if id := item.TelegramInfo.VideoFileID(meta.BotID()); id != "" {
					video = telegoutil.MediaVideo(telegoutil.FileFromID(id))
				} else {
					storDetail := v.GetOriginalStorage()
					if storDetail != shared.ZeroStorageDetail {
						file, err := serv.StorageGetFile(ctx, storDetail)
						if err != nil {
							return oops.Wrapf(err, "failed to get file from storage")
						}
						defer file.Close()
						videoFile, err := osutil.OpenTemp(file.Name())
						if err != nil {
							return oops.Wrapf(err, "failed to open video file")
						}
						video = telegoutil.MediaVideo(telegoutil.File(videoFile))
						closers = append(closers, func() error { return videoFile.Close() })
					} else {
						file, err := httpclient.DownloadWithCache(ctx, v.GetURL(), nil)
						if err != nil {
							return oops.Wrapf(err, "failed to download file: %s", v.GetURL())
						}
						defer file.Close()
						videoFile, err := osutil.OpenTemp(file.Name())
						if err != nil {
							return oops.Wrapf(err, "failed to open video file")
						}
						video = telegoutil.MediaVideo(telegoutil.File(videoFile))
						closers = append(closers, func() error { return videoFile.Close() })
					}
				}
				if video == nil {
					return oops.New("failed to create input media video")
				}
				if i == start {
					video = video.WithCaption(caption).WithParseMode(telego.ModeHTML)
				}
				if artwork.GetR18() {
					video = video.WithHasSpoiler()
				}
				video.WithSupportsStreaming()
				rs, ok := video.Media.File.(io.ReadSeeker)
				if ok {
					// extract video metadata
					var meta *mediatool.VideoMetadata
					if mediatool.FFmpegAvailable() {
						meta, _ = mediatool.GetVideoMetadata(rs)
					} else {
						meta, _ = mediatool.GetMP4Meta(rs)
					}
					if meta != nil {
						video = video.WithWidth(int(meta.Width)).WithHeight(int(meta.Height)).WithDuration(int(meta.Duration / 1000))
					}
					// extract video cover
					if mediatool.FFmpegAvailable() {
						rs.Seek(0, io.SeekStart)
						thumb, err := mediatool.ExtractVideoThumbFrame(rs)
						if err == nil {
							cover := telegoutil.FileFromBytes(thumb, "thumb.jpg")
							video = video.WithCover(&cover)
						} else {
							log.Warnf("failed to extract video thumb frame: %v", err)
						}
					}
					rs.Seek(0, io.SeekStart)
				}
				inputMedia = video
			}

			inputMedias[i-start] = inputMedia
			return nil
		}()
		if err != nil {
			var closeErrs []error
			for _, closer := range closers {
				if err := closer(); err != nil {
					closeErrs = append(closeErrs, err)
				}
			}
			return nil, oops.Wrapf(err, "failed to create input media, close errs: %v", oops.Join(closeErrs...))
		}
	}

	return &ioutil.Closer[[]telego.InputMedia]{
		Value: inputMedias,
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
