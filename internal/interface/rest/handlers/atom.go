package handlers

import (
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gorilla/feeds"
	"github.com/krau/ManyACG/internal/infra/config/runtimecfg"
	"github.com/krau/ManyACG/internal/interface/rest/common"
	"github.com/krau/ManyACG/internal/interface/rest/utils"
	"github.com/krau/ManyACG/internal/model/query"
	"github.com/krau/ManyACG/internal/service"
	"github.com/krau/ManyACG/internal/shared"
)

func GenerateAtomFeed(ctx fiber.Ctx) error {
	serv := common.MustGetState[*service.Service](ctx, common.StateKeyService)
	cfg := common.MustGetState[runtimecfg.RestConfig](ctx, common.StateKeyConfig)

	artworks, err := serv.QueryArtworks(ctx, query.ArtworksDB{
		// 默认排序即为发布时间降序
		Paginate: query.Paginate{
			Limit: 50,
		},
		ArtworksFilter: query.ArtworksFilter{
			R18: shared.R18TypeNone,
		},
	})
	if err != nil {
		return common.NewError(fiber.StatusInternalServerError, " failed to query artworks")
	}
	feed := &feeds.Feed{
		Title:       cfg.Site.Title,
		Link:        &feeds.Link{Href: cfg.Site.URL},
		Description: cfg.Site.Desc,
		Author:      &feeds.Author{Name: cfg.Site.Name, Email: cfg.Site.Email},
		Created:     time.Now(),
		Items:       utils.EntityArtworkToFeedItems(cfg, artworks),
	}
	atom, err := feed.ToAtom()
	if err != nil {
		return common.NewError(fiber.StatusInternalServerError, "failed to generate atom feed")
	}
	ctx.Set(fiber.HeaderContentType, fiber.MIMEApplicationXML)
	ctx.Send([]byte(atom))
	return nil
}
