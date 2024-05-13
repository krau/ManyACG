package restful

import (
	"ManyACG/config"
	"net/http"

	"github.com/gin-gonic/gin"
)

func AuthRequired(ctx *gin.Context) {
	if ctx.GetHeader("Authorization") != "Bearer "+config.Cfg.API.Token {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"message": "Unauthorized",
		})
		ctx.Abort()
		return
	}
	ctx.Next()
}
