package pixiv

import (
	"context"
	"encoding/xml"
	"fmt"
	"regexp"
	"strings"

	"github.com/duke-git/lancet/v2/slice"
	"github.com/duke-git/lancet/v2/strutil"
	"github.com/duke-git/lancet/v2/validator"
	"github.com/imroc/req/v3"
	"github.com/krau/ManyACG/internal/model/dto"
	"github.com/krau/ManyACG/internal/shared"
	"github.com/samber/oops"
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
	IllustType int    `json:"illustType"` // 2: ugoira (动图) 0: 普通图片
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
	En string `json:"en"` // 实际上会出现其他语言翻译
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

type PixivUgoiraMeta struct {
	Error   bool                 `json:"error"`
	Message string               `json:"message"`
	Body    *PixivUgoiraMetaBody `json:"body"`
}

type PixivUgoiraMetaBody struct {
	Src         string                 `json:"src"`
	OriginalSrc string                 `json:"originalSrc"`
	MimeType    string                 `json:"mimeType"`
	Frames      []PixivUgoiraMetaFrame `json:"frames"`
}

type PixivUgoiraMetaFrame struct {
	File  string `json:"file"`
	Delay int    `json:"delay"`
}

var (
	tagsSet             = map[string]bool{"R-18": true, "R-18G": true, "R18": true, "R18G": true}
	bookmarksTagsSuffix = []string{"入り", "bookmarks", "0收藏", "+ users", "加入书籤"}
	htmlRe              = regexp.MustCompile("<[^>]+>")
)

func (resp *PixivAjaxResp) ToArtwork(
	ctx context.Context,
	client *req.Client,
	imgProxy string,
) (*dto.FetchedArtwork, error) {
	if resp.Err {
		return nil, fmt.Errorf("pixiv ajax response error: %s", resp.Message)
	}
	illustPages, err := reqIllustPages(ctx, "https://www.pixiv.net/artworks/"+resp.Body.IllustId, client)
	if err != nil {
		return nil, oops.Wrapf(err, "request pixiv illust pages error")
	}
	if illustPages.Err {
		return nil, fmt.Errorf("pixiv illust pages response error: %s", illustPages.Message)
	}

	pictures := make([]*dto.FetchedPicture, 0)
	for i, page := range illustPages.Body {
		pictures = append(pictures, &dto.FetchedPicture{
			Index:     uint(i),
			Thumbnail: strings.Replace(page.Urls.Small, pixivImgDomain, imgProxy, 1),
			Original:  strings.Replace(page.Urls.Original, pixivImgDomain, imgProxy, 1),
			Width:     uint(page.Width),
			Height:    uint(page.Height),
		})
	}

	tags := make([]string, 0)
	for _, tag := range resp.Body.Tags.Tags {
		if tag.Translation != nil && tag.Translation.En != "" && validator.ContainChinese(tag.Translation.En) {
			tags = append(tags, tag.Translation.En)
		} else {
			tags = append(tags, tag.Tag)
		}
	}
	r18 := false
	for _, tag := range tags {
		if tagsSet[tag] {
			r18 = true
			break
		}
	}

	tags = slice.Compact(slice.Filter(tags, func(index int, item string) bool {
		return !strutil.HasSuffixAny(item, bookmarksTagsSuffix)
	}))

	fetched := &dto.FetchedArtwork{
		Title:       resp.Body.IlustTitle,
		Description: htmlRe.ReplaceAllString(strings.ReplaceAll(resp.Body.Description, "<br />", "\n"), ""),
		R18:         r18,
		SourceType:  shared.SourceTypePixiv,
		SourceURL:   "https://www.pixiv.net/artworks/" + resp.Body.IllustId,
		Artist: &dto.FetchedArtist{
			Name:     resp.Body.Username,
			Type:     shared.SourceTypePixiv,
			UID:      resp.Body.UserId,
			Username: resp.Body.UserAccount,
		},
		Tags:     tags,
		Pictures: pictures,
	}
	if resp.Body.IllustType == 2 && len(illustPages.Body) == 1 {
		// 动图
		ugMeta, err := reqUgoiraMeta(ctx, fetched.SourceURL, client)
		if err != nil {
			return nil, err
		}
		if ugMeta.Error {
			return nil, oops.Errorf("pixiv ugoira meta response error: %s", ugMeta.Message)
		}
		illustBody := illustPages.Body[0]
		frames := make([]shared.UgoiraFrame, 0)
		for _, frame := range ugMeta.Body.Frames {
			frames = append(frames, shared.UgoiraFrame{
				File:  frame.File,
				Delay: frame.Delay,
			})
		}

		ugoiraData := &dto.FetchedUgoiraMeta{
			Index: 0,
			Data: shared.UgoiraMetaData{
				PosterOriginal: strings.Replace(illustBody.Urls.Original, pixivImgDomain, imgProxy, 1),
				PosterThumb:    strings.Replace(illustBody.Urls.Small, pixivImgDomain, imgProxy, 1),
				Width:          illustBody.Width,
				Height:         illustBody.Height,
				OriginalZip:    strings.Replace(ugMeta.Body.OriginalSrc, pixivImgDomain, imgProxy, 1),
				ThumbZip:       strings.Replace(ugMeta.Body.Src, pixivImgDomain, imgProxy, 1),
				MimeType:       ugMeta.Body.MimeType,
				Frames:         frames,
			},
		}
		fetched.UgoiraMetas = make([]*dto.FetchedUgoiraMeta, 0)
		fetched.UgoiraMetas = append(fetched.UgoiraMetas, ugoiraData)
	}
	return fetched, nil
}
