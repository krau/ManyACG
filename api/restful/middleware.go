package restful

import (
	"ManyACG/config"
	"net/http"

	"github.com/gin-gonic/gin"
)

func AuthRequired(ctx *gin.Context) {
	authHeader := ctx.GetHeader("Authorization")
	tokenQuery := ctx.DefaultQuery("token", "")
	if authHeader == "Bearer "+config.Cfg.API.Token || tokenQuery == config.Cfg.API.Token {
		ctx.Next()
		return
	}
	ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
	ctx.Abort()
}
