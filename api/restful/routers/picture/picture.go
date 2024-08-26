package picture

import (
	"ManyACG/api/restful/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterRouter(r *gin.RouterGroup) {
	r.GET("/file/:id", middleware.ValidatePictureID, GetFile)
}
