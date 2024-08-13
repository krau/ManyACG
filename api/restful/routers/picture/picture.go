package picture

import "github.com/gin-gonic/gin"

func RegisterRouter(r *gin.RouterGroup) {
	r.GET("/thumb/:id", ValidatePictureID, GetThumb)
	r.GET("/file/:id", ValidatePictureID, GetFile)
}
