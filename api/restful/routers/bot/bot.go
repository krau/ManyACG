package bot

import "github.com/gin-gonic/gin"

func RegisterRouter(r *gin.RouterGroup) {
	r.POST("/send_artwork_info", SendArtworkInfo)
}
