package restful

import (
	"ManyACG/api/restful/middleware"
	"ManyACG/api/restful/routers"
	"ManyACG/config"
	. "ManyACG/logger"
	"fmt"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func Run() {
	if config.Cfg.Debug {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()

	corsConfig := cors.DefaultConfig()
	corsConfig.AllowCredentials = true
	corsConfig.AllowHeaders = append(corsConfig.AllowHeaders, "Authorization")
	if config.Cfg.Debug {
		fmt.Println("Allowing all origins in debug mode")
		corsConfig.AllowAllOrigins = true
	} else {
		corsConfig.AllowOrigins = config.Cfg.API.AllowedOrigins
	}

	r.Use(cors.New(corsConfig))

	middleware.Init()
	v1 := r.Group("/api/v1")
	routers.RegisterAllRouters(v1, middleware.JWTAuthMiddleware)

	if err := r.Run(config.Cfg.API.Address); err != nil {
		Logger.Fatalf("Failed to start server: %v", err)
		os.Exit(1)
	}
}
