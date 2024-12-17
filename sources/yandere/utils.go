package yandere

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/krau/ManyACG/types"
)

var (
	yandereSourceURLRegexp = regexp.MustCompile(`yande\.re/post/show/\d+`)
	fakeArtist             = &types.Artist{
		Name:     "Yandere",
		Username: "Yandere",
		UID:      "1",
		Type:     types.SourceTypeYandere,
	}
	sourceURLPrefix = "https://yande.re/post/show/"
	apiBaseURL      = "https://yande.re/post.json?tags=id:"
)

var (
	ErrInvalidYanderePostURL = errors.New("invalid yandere post url")
	ErrYandereAPIError       = errors.New("yandere api error")
)

func GetPostID(url string) string {
	matchUrl := yandereSourceURLRegexp.FindString(url)
	if matchUrl == "" {
		return ""
	}
	return strings.Split(matchUrl, "/")[len(strings.Split(matchUrl, "/"))-1]
}

type YandereJsonResp []YanderePostJsonResp

type YanderePostJsonResp struct {
	ID        int    `json:"id"`
	Tags      string `json:"tags"`
	Author    string `json:"author"`
	Source    string `json:"source"`     // title, maybe?
	FileURL   string `json:"file_url"`   // original
	SampleURL string `json:"sample_url"` // thumbnail
	Width     int    `json:"width"`
	Height    int    `json:"height"`
}

func (resp *YandereJsonResp) ToArtwork() *types.Artwork {
	post := (*resp)[0]
	tags := strings.Split(post.Tags, " ")
	pictures := make([]*types.Picture, 0)
	pictures = append(pictures, &types.Picture{
		Index:     0,
		Thumbnail: post.SampleURL,
		Original:  post.FileURL,
		Width:     uint(post.Width),
		Height:    uint(post.Height),
	})
	var title string
	if post.Source != "" {
		title = post.Source
	} else if post.Author != "" {
		title = fmt.Sprintf("%s/%d", post.Author, post.ID)
	} else {
		title = fmt.Sprintf("Yandere/%d", post.ID)
	}
	artwork := &types.Artwork{
		Title:       title,
		Description: "",
		R18:         false,
		SourceType:  types.SourceTypeYandere,
		SourceURL:   sourceURLPrefix + strconv.Itoa(post.ID),
		Artist:      fakeArtist,
		Tags:        tags,
		Pictures:    pictures,
	}
	return artwork
}
