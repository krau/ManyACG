package bot

import (
	"github.com/gin-gonic/gin"
	"github.com/krau/ManyACG/api/restful/middleware"
)

func RegisterRouter(r *gin.RouterGroup) {
	r.Use(middleware.AdminKeyRequired)
	r.POST("/send_artwork_info", SendArtworkInfo)
}
