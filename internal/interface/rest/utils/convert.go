package utils

import (
	"fmt"
	"html"
	"math/rand"
	"net/url"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/gorilla/feeds"
	"github.com/krau/ManyACG/internal/infra/cache"
	"github.com/krau/ManyACG/internal/infra/config/runtimecfg"
	"github.com/krau/ManyACG/internal/model/entity"
	"github.com/krau/ManyACG/internal/shared"
	"github.com/krau/ManyACG/pkg/log"
)

func EntityArtworkToFeedItems(
	ctx fiber.Ctx,
	cfg runtimecfg.RestConfig,
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
					picUrl := ResponseUrlForStoragePath(ctx, *pic.StorageInfo.Data().Regular, cfg.StoragePathRules)
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

// 在api返回中重写存储路径, 用于拼接直链
//
// 例: rule.Type = "alist", rule.Path = "/pictures/", rule.JoinPrefix = "https://example.com/pictures/", rule.TrimPrefix = "/pictures/"
//
// -> https://example.com/pictures/1234567890abcdef.jpg
//
// 如果图片路径以rule.Path开头, 且rule.StorageType为空或与图片存储类型匹配, 则将图片路径转换为rule.JoinPrefix + 图片路径去掉rule.TrimPrefix的部分
//
// 否则空
func ResponseUrlForStoragePath(ctx fiber.Ctx, detail shared.StorageDetail, rules []runtimecfg.StoragePathRule) string {
	cacheKey := fmt.Sprintf("rest:storage_path_rule_result:%s", detail.Hash())
	if v, err := cache.Get[string](ctx, cacheKey); err == nil && v != "" {
		return v
	}
	for _, rule := range rules {
		if strings.HasPrefix(detail.Path, rule.MatchPrefix) && (rule.StorageType == "" || rule.StorageType == detail.Type.String()) {
			prefix := rule.JoinPrefix[rand.Intn(len(rule.JoinPrefix))]
			parsedUrl, err := url.JoinPath(prefix, strings.TrimPrefix(detail.Path, rule.TrimPrefix))
			if err != nil {
				log.Errorf("failed to join path: %v", err)
				return ""
			}
			cache.Set(ctx, cacheKey, parsedUrl)
			return parsedUrl
		}
	}
	return ""
}

// 根据配置生成图片的访问 URL
//
// 返回值一定不为空, 无可用配置时会回落到 pic.Thumbnail
func PictureResponseUrl(ctx fiber.Ctx, pic *entity.Picture, cfg runtimecfg.RestConfig) (thumbnail, regular string) {
	data := pic.StorageInfo.Data()
	if data == shared.ZeroStorageInfo {
		thumbnail = pic.Thumbnail
		regular = pic.Thumbnail
		return
	}
	if data.Thumb != nil {
		thumbnail = ResponseUrlForStoragePath(ctx, *data.Thumb, cfg.StoragePathRules)
		if thumbnail == "" && cfg.Base != "" {
			thumbnail = fmt.Sprintf("%s/picture/file/thumb/%s", strings.TrimRight(cfg.Base, "/"), pic.ID.Hex())
		}
	}
	if data.Regular != nil {
		regular = ResponseUrlForStoragePath(ctx, *data.Regular, cfg.StoragePathRules)
		if regular == "" && cfg.Base != "" {
			regular = fmt.Sprintf("%s/picture/file/regular/%s", strings.TrimRight(cfg.Base, "/"), pic.ID.Hex())
		}
	}
	if thumbnail == "" {
		thumbnail = pic.Thumbnail
	}
	if regular == "" {
		regular = pic.Thumbnail
	}
	return
}
