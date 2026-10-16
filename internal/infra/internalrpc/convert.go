package internalrpc

import (
	"errors"
	"strings"

	"github.com/krau/ManyACG/internal/infra/internalrpc/internalpb"
	"github.com/krau/ManyACG/internal/model/dto"
	"github.com/krau/ManyACG/internal/shared"
	"github.com/krau/ManyACG/internal/shared/errs"
)

func artworkToProto(artwork *dto.FetchedArtwork) *internalpb.FetchedArtwork {
	if artwork == nil {
		return nil
	}
	message := &internalpb.FetchedArtwork{
		Title:       artwork.Title,
		Description: artwork.Description,
		SourceType:  string(artwork.SourceType),
		SourceUrl:   artwork.SourceURL,
		Tags:        artwork.Tags,
		R18:         artwork.R18,
	}
	if artwork.Artist != nil {
		message.Artist = &internalpb.FetchedArtist{
			Name:     artwork.Artist.Name,
			Type:     string(artwork.Artist.Type),
			Uid:      artwork.Artist.UID,
			Username: artwork.Artist.Username,
		}
	}
	for _, picture := range artwork.Pictures {
		if picture == nil {
			continue
		}
		message.Pictures = append(message.Pictures, &internalpb.FetchedPicture{
			Thumbnail: picture.Thumbnail,
			Original:  picture.Original,
			Index:     uint32(picture.Index),
			Width:     uint32(picture.Width),
			Height:    uint32(picture.Height),
		})
	}
	for _, meta := range artwork.UgoiraMetas {
		if meta == nil {
			continue
		}
		ugoira := &internalpb.FetchedUgoiraMeta{
			Index: uint32(meta.Index),
			Data: &internalpb.FetchedUgoiraData{
				PosterOriginal: meta.Data.PosterOriginal,
				PosterThumb:    meta.Data.PosterThumb,
				OriginalZip:    meta.Data.OriginalZip,
				ThumbZip:       meta.Data.ThumbZip,
				MimeType:       meta.Data.MimeType,
				Width:          int32(meta.Data.Width),
				Height:         int32(meta.Data.Height),
			},
		}
		for _, frame := range meta.Data.Frames {
			ugoira.Data.Frames = append(ugoira.Data.Frames, &internalpb.FetchedUgoiraFrame{
				File:  frame.File,
				Delay: int32(frame.Delay),
			})
		}
		message.UgoiraMetas = append(message.UgoiraMetas, ugoira)
	}
	for _, video := range artwork.Videos {
		if video == nil {
			continue
		}
		message.Videos = append(message.Videos, &internalpb.FetchedVideo{
			Url:      video.URL,
			Poster:   video.Poster,
			MimeType: video.MimeType,
			Index:    uint32(video.Index),
			Width:    uint32(video.Width),
			Height:   uint32(video.Height),
			Duration: uint32(video.Duration),
		})
	}
	return message
}

func artworkFromProto(message *internalpb.FetchedArtwork) *dto.FetchedArtwork {
	if message == nil {
		return nil
	}
	artwork := &dto.FetchedArtwork{
		Title:       message.GetTitle(),
		Description: message.GetDescription(),
		SourceType:  shared.SourceType(message.GetSourceType()),
		SourceURL:   message.GetSourceUrl(),
		Tags:        message.GetTags(),
		R18:         message.GetR18(),
	}
	if artist := message.GetArtist(); artist != nil {
		artwork.Artist = &dto.FetchedArtist{
			Name:     artist.GetName(),
			Type:     shared.SourceType(artist.GetType()),
			UID:      artist.GetUid(),
			Username: artist.GetUsername(),
		}
	}
	for _, picture := range message.GetPictures() {
		artwork.Pictures = append(artwork.Pictures, &dto.FetchedPicture{
			Thumbnail: picture.GetThumbnail(),
			Original:  picture.GetOriginal(),
			Index:     uint(picture.GetIndex()),
			Width:     uint(picture.GetWidth()),
			Height:    uint(picture.GetHeight()),
		})
	}
	for _, meta := range message.GetUgoiraMetas() {
		ugoira := &dto.FetchedUgoiraMeta{
			Index: uint(meta.GetIndex()),
			Data: shared.UgoiraMetaData{
				PosterOriginal: meta.GetData().GetPosterOriginal(),
				PosterThumb:    meta.GetData().GetPosterThumb(),
				OriginalZip:    meta.GetData().GetOriginalZip(),
				ThumbZip:       meta.GetData().GetThumbZip(),
				MimeType:       meta.GetData().GetMimeType(),
				Width:          int(meta.GetData().GetWidth()),
				Height:         int(meta.GetData().GetHeight()),
			},
		}
		for _, frame := range meta.GetData().GetFrames() {
			ugoira.Data.Frames = append(ugoira.Data.Frames, shared.UgoiraFrame{
				File:  frame.GetFile(),
				Delay: int(frame.GetDelay()),
			})
		}
		artwork.UgoiraMetas = append(artwork.UgoiraMetas, ugoira)
	}
	for _, video := range message.GetVideos() {
		artwork.Videos = append(artwork.Videos, &dto.FetchedVideo{
			URL:      video.GetUrl(),
			Poster:   video.GetPoster(),
			MimeType: video.GetMimeType(),
			Index:    uint(video.GetIndex()),
			Width:    uint(video.GetWidth()),
			Height:   uint(video.GetHeight()),
			Duration: uint(video.GetDuration()),
		})
	}
	return artwork
}

func failureToError(failure *internalpb.JobFailure) error {
	if failure == nil {
		return errors.New("internal agent job failed")
	}
	if failure.GetCode() == internalpb.FailureCode_FAILURE_CODE_UNSUPPORTED_SOURCE {
		return errs.ErrUnsupportedSource
	}
	if message := strings.TrimSpace(failure.GetMessage()); message != "" {
		return errors.New(message)
	}
	return errors.New("internal agent job failed")
}

func failureFromError(err error) *internalpb.JobFailure {
	code := internalpb.FailureCode_FAILURE_CODE_INTERNAL
	if errors.Is(err, errs.ErrUnsupportedSource) {
		code = internalpb.FailureCode_FAILURE_CODE_UNSUPPORTED_SOURCE
	}
	return &internalpb.JobFailure{Code: code, Message: err.Error()}
}
