package user

import (
	"ManyACG/api/restful/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterRouter(r *gin.RouterGroup) {
	r.GET("/unauth/:id", middleware.ValidateParamObjectID, GetUnauthUser)
}
