package routers

import (
	"ManyACG/api/restful/routers/artwork"
	"ManyACG/api/restful/routers/bot"

	"github.com/gin-gonic/gin"
)

func RegisterAllRouters(r *gin.RouterGroup) {
	r.Use(CheckAuth)

	artworkGroup := r.Group("/artwork")
	artwork.RegisterRouter(artworkGroup)

	botGroup := r.Group("/bot")
	botGroup.Use(AuthRequired)
	bot.RegisterRouter(botGroup)
}
