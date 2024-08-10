package restful

import (
	"ManyACG/api/restful/routers"
	"ManyACG/config"
	. "ManyACG/logger"
	"os"

	"github.com/gin-gonic/gin"
)

func Run() {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	v1 := r.Group("/v1")

	routers.RegisterAllRouters(v1)

	if err := r.Run(config.Cfg.API.Address); err != nil {
		Logger.Fatalf("Failed to start server: %v", err)
		os.Exit(1)
	}
}
