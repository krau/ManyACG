package middleware

import (
	"time"

	"github.com/krau/ManyACG/config"
	. "github.com/krau/ManyACG/logger"
	"github.com/krau/ManyACG/model"
	"github.com/krau/ManyACG/service"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type Login struct {
	Username   string `json:"username" binding:"min=4,max=20"`
	TelegramID int64  `json:"telegram_id" binding:"omitempty"`
	Email      string `json:"email" binding:"omitempty,email"`
	Password   string `json:"password" binding:"required"`
}

type User struct {
	UserName string
}

const (
	IdentityKey = "id"
)

var JWTAuthMiddleware *jwt.GinJWTMiddleware

func JWTInitParamas() *jwt.GinJWTMiddleware {
	return &jwt.GinJWTMiddleware{
		Realm: func() string {
			if config.Cfg.API.Realm != "" {
				return config.Cfg.API.Realm
			} else {
				return "manyacg"
			}
		}(),
		Key:        []byte(config.Cfg.API.Secret),
		Timeout:    time.Minute * time.Duration(config.Cfg.API.TokenExpire),
		MaxRefresh: time.Hour * time.Duration(config.Cfg.API.RefreshTokenExpire),
		Unauthorized: func(c *gin.Context, code int, message string) {
			c.JSON(code, gin.H{
				"status":  code,
				"message": message,
			})
		},
		TokenLookup: "header:Authorization,cookie:TOKEN",

		// 设置自定义 payload
		PayloadFunc: func(data interface{}) jwt.MapClaims {
			if user, ok := data.(Login); ok {
				return jwt.MapClaims{
					IdentityKey: user.Username,
				}
			}
			return jwt.MapClaims{}
		},

		// 从请求中提取用户信息, 返回值会传递给 Authorizator
		IdentityHandler: func(ctx *gin.Context) interface{} {
			claims := jwt.ExtractClaims(ctx)
			return &User{
				UserName: claims[IdentityKey].(string),
			}
		},

		// 登陆成功
		Authorizator: func(data interface{}, c *gin.Context) bool {
			payloadUser, ok := data.(*User)
			if !ok {
				return false
			}
			user, err := service.GetUserByUsername(c, payloadUser.UserName)
			if err != nil || user == nil || user.Blocked {
				return false
			}
			return true
		},

		// Must return user data as user identifier, it will be stored in Claim Array.
		Authenticator: func(c *gin.Context) (interface{}, error) {
			loginInfo := Login{}
			if err := c.ShouldBind(&loginInfo); err != nil {
				Logger.Errorf("Failed to bind login info: %v", err)
				return "", jwt.ErrMissingLoginValues
			}
			if loginInfo.Username == "" && loginInfo.Email == "" && loginInfo.TelegramID == 0 {
				return nil, jwt.ErrMissingLoginValues
			}

			var user *model.UserModel
			if loginInfo.TelegramID != 0 {
				user, _ = service.GetUserByTelegramID(c, loginInfo.TelegramID)
				if user != nil {
					loginInfo.Username = user.Username
				}
			}
			if loginInfo.Email != "" {
				user, _ = service.GetUserByEmail(c, loginInfo.Email)
				if user != nil {
					loginInfo.Username = user.Username
				}
			}
			if loginInfo.Username != "" {
				user, _ = service.GetUserByUsername(c, loginInfo.Username)
			}
			if user == nil {
				return nil, jwt.ErrFailedAuthentication
			}

			if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(loginInfo.Password)); err != nil {
				return nil, jwt.ErrFailedAuthentication
			}
			return loginInfo, nil
		},
	}
}

/*
用在一些可选登录接口上.

如果登录了, 则会在 ctx 中设置 "logged" 为 true, 并设置 "claims" 为 jwt.MapClaims.
*/
func OptionalJWTMiddleware(ctx *gin.Context) {
	token, err := JWTAuthMiddleware.ParseToken(ctx)
	if err == nil && token.Valid {
		claims := jwt.ExtractClaimsFromToken(token)
		ctx.Set("logged", true)
		ctx.Set("claims", claims)
	} else {
		ctx.Set("logged", false)
	}
	ctx.Next()
}
