package utils

import (
	"fmt"
	"html"
	"net/url"
	"strings"

	"github.com/gorilla/feeds"
	"github.com/krau/ManyACG/internal/infra/config/runtimecfg"
	"github.com/krau/ManyACG/internal/model/entity"
	"github.com/krau/ManyACG/internal/shared"
	"github.com/krau/ManyACG/pkg/log"
)

func EntityArtworkToFeedItems(cfg runtimecfg.RestConfig,
	arts []*entity.Artwork) []*feeds.Item {
	if len(arts) == 0 {
		return nil
	}
	items := make([]*feeds.Item, 0, len(arts))
	siteCfg := cfg.Site
	for _, artwork := range arts {
		item := &feeds.Item{
			Title:       artwork.Title,
			Link:        &feeds.Link{Href: fmt.Sprintf("%s/artwork/%s", siteCfg.URL, artwork.ID.Hex())},
			Description: artwork.Description,
			Author:      &feeds.Author{Name: artwork.GetArtist().GetName()},
			Created:     artwork.CreatedAt,
			Updated:     artwork.UpdatedAt,
			Id:          fmt.Sprintf("%s/artwork/%s", siteCfg.URL, artwork.ID.Hex()),
			Content: fmt.Sprintf(`<article><h2>%s</h2><figure><img src="%s" alt="%s" /></figure><p>%s</p><p>Artist: %s</p><p>Created: %s</p></article>`,
				html.EscapeString(artwork.Title),
				html.EscapeString(func() string {
					pic := artwork.Pictures[0]
					if pic.StorageInfo.Data() == shared.ZeroStorageInfo || pic.StorageInfo.Data().Regular == nil {
						return pic.Thumbnail
					}
					picUrl := ApplyApiStoragePathRule(*pic.StorageInfo.Data().Regular, cfg.StoragePathRules)
					if picUrl == "" || picUrl == pic.StorageInfo.Data().Regular.Path {
						return pic.Thumbnail
					}
					return picUrl
				}()),
				html.EscapeString(artwork.Title),
				html.EscapeString(artwork.Description),
				html.EscapeString(artwork.Artist.Name),
				artwork.CreatedAt.Format("2006-01-02 15:04:05")),
		}
		items = append(items, item)
	}
	return items
}

// 在api返回中重写图片路径, 用于拼接图片直链
//
// 例: rule.Type = "alist", rule.Path = "/pictures/", rule.JoinPrefix = "https://example.com/pictures/", rule.TrimPrefix = "/pictures/"
//
// -> https://example.com/pictures/1234567890abcdef.jpg
//
// 如果图片路径以rule.Path开头, 且rule.StorageType为空或与图片存储类型匹配, 则将图片路径转换为rule.JoinPrefix + 图片路径去掉rule.TrimPrefix的部分
//
// 否则空
func ApplyApiStoragePathRule(detail shared.StorageDetail, rules []runtimecfg.StoragePathRule) string {
	for _, rule := range rules {
		if strings.HasPrefix(detail.Path, rule.Path) && (rule.StorageType == "" || rule.StorageType == string(detail.Type)) {
			parsedUrl, err := url.JoinPath(rule.JoinPrefix, strings.TrimPrefix(detail.Path, rule.TrimPrefix))
			if err != nil {
				log.Errorf("failed to join path: %v", err)
				return ""
			}
			return parsedUrl
		}
	}
	return ""
}

func GetPictureResponseUrl(pic *entity.Picture, cfg runtimecfg.RestConfig) (thumbnail, regular string) {
	data := pic.StorageInfo.Data()
	if data == shared.ZeroStorageInfo {
		thumbnail = pic.Thumbnail
		regular = pic.Thumbnail
		return
	}
	if data.Thumb != nil {
		thumbnail = ApplyApiStoragePathRule(*data.Thumb, cfg.StoragePathRules)
		if thumbnail == "" {
			thumbnail = pic.Thumbnail
		}
	}
	if data.Regular != nil {
		regular = ApplyApiStoragePathRule(*data.Regular, cfg.StoragePathRules)
		if regular == "" {
			regular = pic.Thumbnail
		}
	}
	return
}
