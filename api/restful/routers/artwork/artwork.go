package artwork

import "github.com/gin-gonic/gin"

func RegisterRouter(r *gin.RouterGroup) {
	r.GET("/random", RandomArtworks)
	r.POST("/random", RandomArtworks)
	r.GET("/list", GetLatestArtworks)
	r.POST("/list", GetLatestArtworks)
	r.GET("/:id", GetArtwork)
}
