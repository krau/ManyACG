package twitter

import (
	"errors"
	"fmt"
	"strings"

	"github.com/krau/ManyACG/internal/model/dto"
	"github.com/krau/ManyACG/internal/shared"
	"github.com/krau/ManyACG/pkg/strutil"
)

type FxTwitterApiResp struct {
	Tweet   *Tweet `json:"tweet"`
	Message string `json:"message"`
	Code    int    `json:"code"`
}

type Tweet struct {
	Media             *Media `json:"media"`
	Author            Author `json:"author"`
	URL               string `json:"url"`
	ID                string `json:"id"`
	Text              string `json:"text"`
	PossiblySensitive bool   `json:"possibly_sensitive"`
}

type Author struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Username string `json:"screen_name"` // Twitter username
}

type Media struct {
	Photos []MediaItem `json:"photos"`
	Videos []MediaItem `json:"videos"`
}

type MediaItem struct {
	Type         string  `json:"type"`
	URL          string  `json:"url"`                     // Direct link to the media
	Format       string  `json:"format,omitempty"`        // video's mime type or format, e.g. "video/mp4", "gif"
	ThumbnailUrl string  `json:"thumbnail_url,omitempty"` // for videos poster image
	Width        int     `json:"width"`
	Height       int     `json:"height"`
	Duration     float64 `json:"duration,omitempty"` // in seconds, for videos only
}

var (
	ErrInvalidURL    = errors.New("invalid tweet URL")
	ErrIndexOOB      = errors.New("index out of bounds")
	ErrRequestFailed = errors.New("request twitter url failed")
)

func (resp *FxTwitterApiResp) ToArtwork() (*dto.FetchedArtwork, error) {
	if resp.Code != 200 {
		return nil, fmt.Errorf("%w: %s (code: %d)", ErrRequestFailed, resp.Message, resp.Code)
	}
	if resp.Tweet == nil {
		return nil, ErrInvalidURL
	}
	tweet := resp.Tweet
	if tweet.Media == nil {
		return nil, ErrInvalidURL
	}
	media := tweet.Media
	if media == nil || (len(media.Photos) == 0 && len(media.Videos) == 0) {
		return nil, ErrInvalidURL
	}

	pictures := make([]*dto.FetchedPicture, 0)
	for i, photo := range media.Photos {
		picUrl := strings.Split(photo.URL, "?")[0]
		pictures = append(pictures, &dto.FetchedPicture{
			Index:     uint(i),
			Thumbnail: picUrl + "?name=medium",
			Original:  picUrl + "?name=orig",
			Width:     uint(photo.Width),
			Height:    uint(photo.Height),
		})
	}
	videos := make([]*dto.FetchedVideo, 0)
	for i, video := range media.Videos {
		videoUrl := strings.Split(video.URL, "?")[0]
		posterUrl := ""
		if video.ThumbnailUrl != "" {
			posterUrl = strings.Split(video.ThumbnailUrl, "?")[0]
		}
		mime := ""
		switch video.Format {
		case "gif":
			mime = "image/gif"
		default:
			mime = video.Format
		}
		videos = append(videos, &dto.FetchedVideo{
			Index:    uint(i),
			URL:      videoUrl,
			Width:    uint(video.Width),
			Height:   uint(video.Height),
			Duration: uint(video.Duration * 1000), // convert to milliseconds
			Poster:   posterUrl,
			MimeType: mime,
		})
		// if len(pictures) == 0 && posterUrl != "" {
		// 	// use video poster as picture if no pictures available
		// 	// this is for the temporary compatibility as we assume artworks always have pictures
		// 	pictures = append(pictures, &dto.FetchedPicture{
		// 		Index:     uint(i),
		// 		Thumbnail: posterUrl,
		// 		Original:  posterUrl,
		// 	})
		// }
	}

	title := fmt.Sprintf("%s/%s", tweet.Author.Username, tweet.ID)
	tags := strutil.ExtractTagsFromText(tweet.Text)
	desc := tweet.Text

	if tweet.Text != "" {
		textLines := strings.Split(tweet.Text, "\n")
		firstLine := textLines[0]
		if len(firstLine) <= 114 {
			title = firstLine
		}
	}

	return &dto.FetchedArtwork{
		Title:       title,
		Description: desc,
		SourceType:  shared.SourceTypeTwitter,
		SourceURL:   fmt.Sprintf("https://x.com/%s/status/%s", tweet.Author.Username, tweet.ID),
		R18:         tweet.PossiblySensitive,
		Artist: &dto.FetchedArtist{
			Name:     tweet.Author.Name,
			Username: tweet.Author.Username,
			Type:     shared.SourceTypeTwitter,
			UID:      tweet.Author.ID,
		},
		Pictures: pictures,
		Videos:   videos,
		Tags:     tags,
	}, nil
}
