package twitter

import (
	"ManyACG-Bot/types"
	"errors"
	"fmt"
	"strconv"
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
	ErrInvalidURL = errors.New("invalid tweet URL")
	ErrIndexOOB   = errors.New("index out of bounds")
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

	uid, err := strconv.Atoi(tweet.ID)
	if err != nil {
		return nil, err
	}
	tweetPath, err := GetTweetPath(tweet.URL)
	if err != nil {
		return nil, err
	}
	return &types.Artwork{
		Title:       tweetPath,
		Description: tweet.Text,
		SourceType:  types.SourceTypeTwitter,
		SourceURL:   tweet.URL,
		R18:         false, // TODO
		Artist: &types.Artist{
			Name:     tweet.Author.Name,
			Username: tweet.Author.Username,
			Type:     types.SourceTypeTwitter,
			UID:      uid,
		},
		Pictures: pictures,
		Tags:     nil, // TODO
	}, nil
}
