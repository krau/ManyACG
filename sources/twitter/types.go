package twitter

import (
	"errors"
	"fmt"
	"strings"

	"github.com/krau/ManyACG/common"
	"github.com/krau/ManyACG/types"
)

type FxTwitterApiResp struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Tweet   *Tweet `json:"tweet"`
}

type Tweet struct {
	URL    string `json:"url"`
	ID     string `json:"id"`
	Text   string `json:"text"`
	Author Author `json:"author"`
	Media  *Media `json:"media"`
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

func (resp *FxTwitterApiResp) ToArtwork() (*types.Artwork, error) {
	if resp.Code != 200 {
		return nil, errors.New(resp.Message + " (code: " + fmt.Sprint(resp.Code) + ")")
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

	pictures := make([]*types.Picture, 0)
	for i, photo := range media.Photos {
		pictures = append(pictures, &types.Picture{
			Index:     uint(i),
			Thumbnail: photo.URL + "?name=medium",
			Original:  photo.URL + "?name=orig",
			Width:     uint(photo.Width),
			Height:    uint(photo.Height),
		})
	}
	tweetPath := GetTweetPath(tweet.URL)
	if tweetPath == "" {
		return nil, ErrInvalidURL
	}

	title := tweetPath
	tags := common.ExtractTagsFromText(tweet.Text)
	var desc string

	if tweet.Text != "" {
		textLines := strings.Split(tweet.Text, "\n")
		textLineLen := len(textLines)
		firstLine := textLines[0]
		if len(firstLine) <= 114 {
			title = firstLine
			if textLineLen > 1 {
				desc = strings.Join(textLines[1:textLineLen-1], "\n")
			}
		} else {
			desc = strings.Join(textLines[:textLineLen-1], "\n")
		}
	}

	return &types.Artwork{
		Title:       title,
		Description: desc,
		SourceType:  types.SourceTypeTwitter,
		SourceURL:   tweet.URL,
		R18:         false, // TODO
		Artist: &types.Artist{
			Name:     tweet.Author.Name,
			Username: tweet.Author.Username,
			Type:     types.SourceTypeTwitter,
			UID:      tweet.Author.ID,
		},
		Pictures: pictures,
		Tags:     tags,
	}, nil
}
