package restful

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/krau/ManyACG/api/restful/middleware"
	"github.com/krau/ManyACG/api/restful/routers"
	"github.com/krau/ManyACG/common"
	"github.com/krau/ManyACG/config"

	"github.com/penglongli/gin-metrics/ginmetrics"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func Run(ctx context.Context) {
	if config.Cfg.Debug {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()

	if config.Cfg.API.Metrics {
		metrics := ginmetrics.GetMonitor()
		metrics.SetMetricPath("/metrics")
		metrics.Use(r)
	}

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

	server := &http.Server{
		Addr:    config.Cfg.API.Address,
		Handler: r,
	}

	go func() {
		<-ctx.Done()
		common.Logger.Info("Shutting down api server...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			common.Logger.Fatalf("Failed to shutdown server: %v", err)
		}
		common.Logger.Info("API server stopped")
	}()

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			common.Logger.Fatalf("Failed to start server: %v", err)
			os.Exit(1)
		}
	}()
}
