package auth

import (
	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
)

func RegisterRouter(r *gin.RouterGroup, handle *jwt.GinJWTMiddleware) {
	r.POST("/login", handle.LoginHandler)
	r.POST("/logout", handle.LogoutHandler)
	r.POST("/send_code", handleSendCode)
	r.POST("/register", handleRegister)
	auth := r.Group("/auth", handle.MiddlewareFunc())
	auth.GET("/refresh_token", handle.RefreshHandler)
}
