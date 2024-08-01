package restful

import (
	"ManyACG/config"
	. "ManyACG/logger"
	"os"

	"github.com/gin-gonic/gin"
)

func Run() {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	v1 := r.Group("/v1")
	if config.Cfg.API.Auth {
		v1.Use(AuthRequired)
	}

	{
		v1.GET("/ping", Ping)
		v1.GET("/artwork/random", RandomArtwork)
		v1.POST("/artwork/random", RandomArtwork)
		v1.POST("/artwork/send_info", SendArtworkInfo)
	}

	if err := r.Run(config.Cfg.API.Address); err != nil {
		Logger.Fatalf("Failed to start server: %v", err)
		os.Exit(1)
	}
}
