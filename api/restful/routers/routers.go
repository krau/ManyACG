package routers

import (
	cache "github.com/chenyahui/gin-cache"
	"github.com/krau/ManyACG/api/restful/middleware"
	"github.com/krau/ManyACG/api/restful/routers/artist"
	"github.com/krau/ManyACG/api/restful/routers/artwork"
	"github.com/krau/ManyACG/api/restful/routers/auth"
	"github.com/krau/ManyACG/api/restful/routers/bot"
	"github.com/krau/ManyACG/api/restful/routers/picture"
	"github.com/krau/ManyACG/api/restful/routers/tag"
	"github.com/krau/ManyACG/api/restful/routers/user"
	"github.com/krau/ManyACG/config"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
)

func RegisterAllRouters(r *gin.RouterGroup, authMiddleware *jwt.GinJWTMiddleware) {
	r.Use(middleware.CheckKey)

	if config.Cfg.API.MustKey {
		r.Use(middleware.KeyRequired)
	}

	auth.RegisterRouter(r, authMiddleware)

	if middleware.CacheStore != nil {
		r.GET("/atom", cache.CacheByRequestPath(middleware.CacheStore, middleware.GetCacheDuration("/atom")), GenerateAtom)
	} else {
		r.GET("/atom", GenerateAtom)
	}

	artworkGroup := r.Group("/artwork")
	artwork.RegisterRouter(artworkGroup)

	botGroup := r.Group("/bot")
	botGroup.Use(middleware.KeyRequired)
	bot.RegisterRouter(botGroup)

	tagGroup := r.Group("/tag")
	tag.RegisterRouter(tagGroup)

	pictureGroup := r.Group("/picture")
	picture.RegisterRouter(pictureGroup)

	userGroup := r.Group("/user")
	user.RegisterRouter(userGroup)

	artistGroup := r.Group("/artist")
	artist.RegisterRouter(artistGroup)
}
