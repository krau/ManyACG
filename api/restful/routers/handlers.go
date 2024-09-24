package routers

import (
	"ManyACG/adapter"
	"ManyACG/common"
	"ManyACG/config"
	"ManyACG/service"
	"ManyACG/types"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/feeds"
)

func GenerateAtom(ctx *gin.Context) {
	artworks, err := service.GetLatestArtworks(ctx, types.R18TypeNone, 1, 50)
	if err != nil {
		common.GinErrorResponse(ctx, err, http.StatusInternalServerError, "Failed to get artworks")
		return
	}
	feed := &feeds.Feed{
		Title:       config.Cfg.API.SiteTitle,
		Link:        &feeds.Link{Href: config.Cfg.API.SiteURL},
		Description: config.Cfg.API.SiteDescription,
		Author:      &feeds.Author{Name: config.Cfg.API.SiteName, Email: config.Cfg.API.SiteEmail},
		Created:     time.Now(),
		Items:       adapter.ConvertToFeedItems(ctx, artworks),
	}
	atom, err := feed.ToAtom()
	if err != nil {
		common.GinErrorResponse(ctx, err, http.StatusInternalServerError, "Failed to generate Atom feed")
		return
	}
	ctx.Data(http.StatusOK, "application/xml", []byte(atom))
}