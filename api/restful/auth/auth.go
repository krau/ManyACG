package auth

import (
	"ManyACG/model"
	"ManyACG/service"
	"errors"
	"net/http"

	. "ManyACG/logger"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

func RegisterRouter(r *gin.RouterGroup, handle *jwt.GinJWTMiddleware) {
	r.POST("/login", handle.LoginHandler)
	r.POST("/logout", handle.LogoutHandler)
	r.POST("/register", handleRegister)
	auth := r.Group("/auth", handle.MiddlewareFunc())
	auth.GET("/refresh_token", handle.RefreshHandler)
}

type Register struct {
	Username string `json:"username" binding:"required,min=4,max=20"`
	Password string `json:"password" binding:"required,min=6,max=20"`
}

func handleRegister(c *gin.Context) {
	var register Register
	if err := c.ShouldBind(&register); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  http.StatusBadRequest,
			"message": err.Error(),
		})
		return
	}
	user, err := service.GetUserByUsername(c, register.Username)
	if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
		Logger.Errorf("Failed to get user: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  http.StatusInternalServerError,
			"message": "Failed to get user",
		})
		return
	}
	if user != nil {
		c.JSON(http.StatusConflict, gin.H{
			"status":  http.StatusConflict,
			"message": "User already exists",
		})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(register.Password), bcrypt.DefaultCost)
	if err != nil {
		Logger.Errorf("Failed to hash password: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  http.StatusInternalServerError,
			"message": "Failed to hash password",
		})
		return
	}
	_, err = service.CreateUser(c, &model.UserModel{
		Username: register.Username,
		Password: string(hashedPassword),
	})
	if err != nil {
		Logger.Errorf("Failed to create user: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  http.StatusInternalServerError,
			"message": "Failed to create user",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status":  http.StatusOK,
		"message": "User created",
	})
}
