package artist

import (
	"github.com/krau/ManyACG/api/restful/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterRouter(r *gin.RouterGroup) {
	r.GET("/:id", middleware.ValidateParamObjectID, GetArtist)
}
