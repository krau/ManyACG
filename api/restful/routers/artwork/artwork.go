package artwork

import "github.com/gin-gonic/gin"

func RegisterRouter(r *gin.RouterGroup) {
	r.GET("/random", RandomArtwork)
	r.POST("/random", RandomArtwork)
	r.GET("/list", GetLatestArtworks)
	r.POST("/list", GetLatestArtworks)
	r.GET("/:id", GetArtwork)
}
