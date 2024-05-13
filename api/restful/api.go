package restful

import (
	"ManyACG/config"
	. "ManyACG/logger"
	"os"

	"github.com/gin-gonic/gin"
)

func Run() {
	r := gin.Default()

	v1 := r.Group("/v1")
	if config.Cfg.API.Auth {
		v1.Use(AuthRequired)
	}

	{
		v1.GET("/ping", Ping)
		v1.GET("/artwork/random", RandomArtwork)
		v1.POST("/artwork/random", RandomArtwork)
	}

	if r.Run(config.Cfg.API.Address) != nil {
		Logger.Fatal("Failed to start restful API")
		os.Exit(1)
	}
}
