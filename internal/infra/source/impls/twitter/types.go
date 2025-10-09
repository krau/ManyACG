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
	Code    int    `json:"code"`
	Message string `json:"message"`
	Tweet   *Tweet `json:"tweet"`
}

type Tweet struct {
	URL               string `json:"url"`
	ID                string `json:"id"`
	Text              string `json:"text"`
	PossiblySensitive bool   `json:"possibly_sensitive"`
	Author            Author `json:"author"`
	Media             *Media `json:"media"`
}

type Author struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Username string `json:"screen_name"` // Twitter username
}

type Media struct {
	Photos []MediaItem `json:"photos"`
}

type MediaItem struct {
	Type   string `json:"type"`
	URL    string `json:"url"` // Direct link to the media
	Width  int    `json:"width"`
	Height int    `json:"height"`
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
	if len(media.Photos) == 0 {
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
		Tags:     tags,
	}, nil
}
