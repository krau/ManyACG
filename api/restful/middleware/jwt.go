package middleware

import (
	"ManyACG/config"
	. "ManyACG/logger"
	"ManyACG/service"
	"os"
	"time"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type Login struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type User struct {
	UserName string
}

const (
	IdentityKey = "id"
)

var JWTAuthMiddleware *jwt.GinJWTMiddleware

func Init() {
	var err error
	JWTAuthMiddleware, err = jwt.New(JWTInitParamas())
	if err != nil {
		Logger.Fatalf("JWT init error: %v", err)
		os.Exit(1)
	}
	if err := JWTAuthMiddleware.MiddlewareInit(); err != nil {
		Logger.Fatalf("JWT middleware init error: %v", err)
		os.Exit(1)
	}
}

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

		Authenticator: func(c *gin.Context) (interface{}, error) {
			loginInfo := Login{}
			if err := c.ShouldBind(&loginInfo); err != nil {
				return "", jwt.ErrMissingLoginValues
			}

			user, err := service.GetUserByUsername(c, loginInfo.Username)
			if err != nil {
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
