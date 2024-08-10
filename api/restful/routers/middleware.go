package routers

import (
	"ManyACG/config"
	"net/http"

	"github.com/gin-gonic/gin"
)

func CheckAuth(ctx *gin.Context) {
	authHeader := ctx.GetHeader("Authorization")
	tokenQuery := ctx.DefaultQuery("token", "")
	if authHeader == "Bearer "+config.Cfg.API.Token || tokenQuery == config.Cfg.API.Token {
		ctx.Set("auth", true)
		ctx.Next()
		return
	}
	ctx.Set("auth", false)
	ctx.Next()
}

func AuthRequired(ctx *gin.Context) {
	if ctx.GetBool("auth") {
		ctx.Next()
		return
	}
	ctx.JSON(http.StatusUnauthorized, gin.H{
		"message": "Unauthorized",
	})
	ctx.Abort()
}
