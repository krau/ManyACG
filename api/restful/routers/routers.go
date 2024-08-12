package routers

import (
	"ManyACG/api/restful/routers/artwork"
	"ManyACG/api/restful/routers/bot"
	"ManyACG/api/restful/routers/picture"
	"ManyACG/api/restful/routers/tag"

	"github.com/gin-gonic/gin"
)

func RegisterAllRouters(r *gin.RouterGroup) {
	r.Use(CheckAuth)

	artworkGroup := r.Group("/artwork")
	artwork.RegisterRouter(artworkGroup)

	botGroup := r.Group("/bot")
	botGroup.Use(AuthRequired)
	bot.RegisterRouter(botGroup)

	tagGroup := r.Group("/tag")
	tag.RegisterRouter(tagGroup)

	pictureGroup := r.Group("/picture")
	picture.RegisterRouter(pictureGroup)
}
