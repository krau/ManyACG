package routers

import (
	"ManyACG/api/restful/middleware"
	"ManyACG/api/restful/routers/artwork"
	"ManyACG/api/restful/routers/bot"
	"ManyACG/api/restful/routers/picture"
	"ManyACG/api/restful/routers/tag"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
)

func RegisterAllRouters(r *gin.RouterGroup, authMiddleware *jwt.GinJWTMiddleware) {
	r.Use(middleware.CheckKey)

	artworkGroup := r.Group("/artwork")
	artwork.RegisterRouter(artworkGroup)

	botGroup := r.Group("/bot")
	botGroup.Use(middleware.KeyRequired)
	bot.RegisterRouter(botGroup)

	tagGroup := r.Group("/tag")
	tag.RegisterRouter(tagGroup)

	pictureGroup := r.Group("/picture")
	picture.RegisterRouter(pictureGroup)
}
