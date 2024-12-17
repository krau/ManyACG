package yandere

import (
	"cmp"
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/duke-git/lancet/v2/slice"
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
	apiBaseURL      = "https://yande.re/post.json?tags=parent:"
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
	ParentID  int    `json:"parent_id"`
}

func (resp YandereJsonResp) ToArtwork() *types.Artwork {
	slices.SortFunc(resp, func(a, b YanderePostJsonResp) int {
		if a.ParentID == 0 && b.ParentID != 0 {
			return -1
		}
		if a.ParentID != 0 && b.ParentID == 0 {
			return 1
		}
		return cmp.Compare(a.ID, b.ID)
	})

	var pictures []*types.Picture
	var tags []string
	var title, description, sourceURL string
	gotParent := false

	for i, post := range resp {
		if post.ParentID == 0 && !gotParent {
			title = fmt.Sprintf("%s/%d", post.Author, post.ID)
			if title == "" {
				title = fmt.Sprintf("Yandere/%d", post.ID)
			}
			description = post.Source
			sourceURL = sourceURLPrefix + strconv.Itoa(post.ID)
			gotParent = true
		}
		tags = append(tags, strings.Split(post.Tags, " ")...)
		pictures = append(pictures, &types.Picture{
			Index:     uint(i),
			Thumbnail: post.SampleURL,
			Original:  post.FileURL,
			Width:     uint(post.Width),
			Height:    uint(post.Height),
		})
	}

	tags = slice.Union(tags)

	return &types.Artwork{
		Title:       title,
		Description: description,
		R18:         false,
		SourceType:  types.SourceTypeYandere,
		SourceURL:   sourceURL,
		Artist:      fakeArtist,
		Tags:        tags,
		Pictures:    pictures,
	}
}
