package picture

import (
	"ManyACG/api/restful/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterRouter(r *gin.RouterGroup) {
	r.GET("/thumb/:id", middleware.ValidatePictureID, GetThumb)
	r.GET("/file/:id", middleware.JWTAuthMiddleware.MiddlewareFunc(), middleware.ValidatePictureID, GetFile)
}
