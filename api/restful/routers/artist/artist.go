package artist

import (
	"ManyACG/api/restful/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterRouter(r *gin.RouterGroup) {
	r.GET("/:id", middleware.ValidateParamObjectID, GetArtist)
}
