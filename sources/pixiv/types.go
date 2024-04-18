package pixiv

import (
	"ManyACG-Bot/config"
	"ManyACG-Bot/types"
	"encoding/xml"
	"errors"
	"regexp"
	"strconv"
	"strings"
)

type PixivRss struct {
	XMLName xml.Name `xml:"rss"`
	Channel channel  `xml:"channel"`
}

type channel struct {
	XMLName xml.Name `xml:"channel"`
	Items   []Item   `xml:"item"`
}

type Item struct {
	XMLName     xml.Name `xml:"item"`
	Title       string   `xml:"title"`
	Description string   `xml:"description"`
	PubDate     string   `xml:"pubDate"`
	Guid        string   `xml:"guid"`
	Link        string   `xml:"link"`
	Author      string   `xml:"author"`
}

type PixivAjaxResp struct {
	Err     bool               `json:"error"`
	Message string             `json:"message"`
	Body    *PixivAjaxRespBody `json:"body"`
}

type PixivAjaxRespBody struct {
	IllustId   string `json:"illustId"`
	IlustTitle string `json:"illustTitle"`
	IllustType int    `json:"illustType"`
	Urls       struct {
		Mini     string `json:"mini"`
		Thumb    string `json:"thumb"`
		Small    string `json:"small"`
		Regular  string `json:"regular"`
		Original string `json:"original"`
	} `json:"urls"`
	Tags        PixivAjaxRespBodyTags `json:"tags"`
	UserId      string                `json:"userId"`
	Username    string                `json:"userName"`
	UserAccount string                `json:"userAccount"`
	Description string                `json:"description"`
}

type PixivAjaxRespBodyTags struct {
	Tags []PixivAjaxRespBodyTagsTag `json:"tags"`
}

type PixivAjaxRespBodyTagsTag struct {
	// 返回里确实就是这么套的
	Tag         string                           `json:"tag"`
	Translation *PixivAjaxRespBodyTagTranslation `json:"translation"`
}

type PixivAjaxRespBodyTagTranslation struct {
	En string `json:"en"`
}

type PixivIllustPages struct {
	Err     bool                    `json:"error"`
	Message string                  `json:"message"`
	Body    []*PixivIllustPagesBody `json:"body"`
}

type PixivIllustPagesBody struct {
	Urls struct {
		ThumbMini string `json:"thumb_mini"`
		Small     string `json:"small"`
		Regular   string `json:"regular"`
		Original  string `json:"original"`
	} `json:"urls"`
	Width  int `json:"width"`
	Height int `json:"height"`
}

var (
	tagsSet = map[string]bool{"R-18": true, "R-18G": true, "R18": true, "R18G": true}
	htmlRe  = regexp.MustCompile("<[^>]+>")
)

func (resp *PixivAjaxResp) ToArtwork() (*types.Artwork, error) {
	if resp.Err {
		return nil, errors.New(resp.Message)
	}
	illustPages, err := reqIllustPages("https://www.pixiv.net/artworks/" + resp.Body.IllustId)
	if err != nil {
		return nil, err
	}
	if illustPages.Err {
		return nil, errors.New(illustPages.Message)
	}

	pictures := make([]*types.Picture, 0)
	for i, page := range illustPages.Body {
		pictures = append(pictures, &types.Picture{
			Index:     uint(i),
			Thumbnail: strings.Replace(page.Urls.Small, "i.pximg.net", config.Cfg.Source.Pixiv.Proxy, 1),
			Original:  strings.Replace(page.Urls.Original, "i.pximg.net", config.Cfg.Source.Pixiv.Proxy, 1),
			Width:     uint(page.Width),
			Height:    uint(page.Height),
		})
	}

	tags := make([]*types.ArtworkTag, 0)
	for _, tag := range resp.Body.Tags.Tags {
		tags = append(tags, &types.ArtworkTag{Name: tag.Tag})
	}
	r18 := false
	for _, tag := range tags {
		if tagsSet[tag.Name] {
			r18 = true
		}
	}
	uid, err := strconv.Atoi(resp.Body.UserId)
	if err != nil {
		return nil, err
	}
	return &types.Artwork{
		Title:       resp.Body.IlustTitle,
		Description: htmlRe.ReplaceAllString(resp.Body.Description, ""),
		R18:         r18,
		Source: types.ArtworkSource{
			Type: types.SourceTypePixiv,
			URL:  "https://www.pixiv.net/artworks/" + resp.Body.IllustId,
		},
		Artist: types.Artist{
			Name:     resp.Body.Username,
			Type:     types.SourceTypePixiv,
			UID:      uid,
			Username: resp.Body.UserAccount,
		},
		Tags:     tags,
		Pictures: pictures,
	}, nil
}
