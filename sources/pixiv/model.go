package pixiv

import (
	"ManyACG-Bot/model"
	"encoding/xml"
	"regexp"
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

var (
	tagsSet = map[string]bool{"R-18": true, "R-18G": true, "R18": true, "R18G": true}
	htmlRe  = regexp.MustCompile("<[^>]+>")
)

func (item *Item) ToArtwork(artworkInfo *PixivAjaxResp) *model.Artwork {
	imgs := strings.Split(item.Description, "<img src=\"")
	srcs := make([]string, 0)
	for _, img := range imgs {
		if strings.HasPrefix(img, "http") {
			src := strings.Split(img, "\"")[0]
			srcs = append(srcs, src)
		}
	}
	pictures := make([]*model.Picture, 0)
	for _, src := range srcs {
		picture := model.Picture{
			DirectURL: src,
		}
		pictures = append(pictures, &picture)
	}

	tags := make([]string, 0)
	for _, tag := range artworkInfo.Body.Tags.Tags {
		var tagName string
		if tag.Translation != nil {
			tagName = tag.Translation.En
		} else {
			tagName = tag.Tag
		}
		tags = append(tags, tagName)
	}
	isR18 := false

	for _, tag := range tags {
		if _, ok := tagsSet[tag]; ok {
			isR18 = true
			break
		}
	}
	artwork := model.Artwork{
		Title:       item.Title,
		Author:      item.Author,
		Description: htmlRe.ReplaceAllString(artworkInfo.Body.Description, ""),
		SourceType:  "pixiv",
		SourceURL:   item.Link,
		Tags:        tags,
		R18:         isR18,
		Pictures:    pictures,
	}
	return &artwork
}

type PixivAjaxResp struct {
	Err     bool               `json:"error"`
	Message string             `json:"message"`
	Body    *PixivAjaxRespBody `json:"body"`
}

type PixivAjaxRespBody struct {
	IllustId    string                `json:"illustId"`
	IllustType  int                   `json:"illustType"`
	Tags        PixivAjaxRespBodyTags `json:"tags"`
	UserId      string                `json:"userId"`
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
	// en翻译实际上是中文
	En string `json:"en"`
}
