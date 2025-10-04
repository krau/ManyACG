package artwork

import (
	"net/http"

	cache "github.com/chenyahui/gin-cache"
	"github.com/krau/ManyACG/api/restful/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterRouter(r *gin.RouterGroup) {
	if middleware.CacheStore != nil {
		registerRoutesWithCache(r)
		registerNoCacheRoutes(r)
		return
	}
	r.Match([]string{http.MethodGet, http.MethodPost},
		"/random",
		middleware.OptionalJWTMiddleware,
		RandomArtworks)
	r.Match([]string{http.MethodGet, http.MethodPost},
		"/list",
		middleware.OptionalJWTMiddleware,
		GetArtworkList)
	r.GET("/count", GetArtworkCount)
	r.Match([]string{http.MethodGet, http.MethodPost},
		"/fetch",
		middleware.CheckApiKey,
		checkApiKeyFetchArtworkPermission,
		FetchArtwork)
	r.GET("/:id",
		middleware.OptionalJWTMiddleware,
		GetArtwork)
	registerNoCacheRoutes(r)
}

func registerRoutesWithCache(r *gin.RouterGroup) {
	r.Match([]string{http.MethodGet, http.MethodPost},
		"/random",
		cache.CacheByRequestURI(middleware.CacheStore, middleware.GetCacheDuration("/artwork/random"), cache.IgnoreQueryOrder()),
		middleware.OptionalJWTMiddleware,
		RandomArtworks)
	r.Match([]string{http.MethodGet, http.MethodPost},
		"/list",
		cache.CacheByRequestURI(middleware.CacheStore, middleware.GetCacheDuration("/artwork/list"), cache.IgnoreQueryOrder()),
		middleware.OptionalJWTMiddleware,
		GetArtworkList)
	r.GET("/count",
		cache.CacheByRequestURI(middleware.CacheStore, middleware.GetCacheDuration("/artwork/count")),
		GetArtworkCount)
	r.Match([]string{http.MethodGet, http.MethodPost},
		"/fetch",
		cache.CacheByRequestURI(middleware.CacheStore, middleware.GetCacheDuration("/artwork/fetch")),
		middleware.CheckApiKey,
		checkApiKeyFetchArtworkPermission,
		FetchArtwork)
	r.GET("/:id",
		cache.CacheByRequestPath(middleware.CacheStore, middleware.GetCacheDuration("/artwork/:id")),
		middleware.OptionalJWTMiddleware,
		GetArtwork)
}

func registerNoCacheRoutes(r *gin.RouterGroup) {
	r.Match([]string{http.MethodGet, http.MethodPost},
		"/random/preview",
		middleware.OptionalJWTMiddleware,
		RandomArtworkPreview)
	r.GET("/like",
		middleware.JWTAuthMiddleware.MiddlewareFunc(),
		validateArtworkIDMiddleware,
		checkArtworkAndUserMiddleware,
		GetArtworkLikeStatus)
	r.POST("/like",
		middleware.JWTAuthMiddleware.MiddlewareFunc(),
		validateArtworkIDMiddleware,
		checkArtworkAndUserMiddleware,
		LikeArtwork)
	r.GET("/favorite",
		middleware.JWTAuthMiddleware.MiddlewareFunc(),
		validateArtworkIDMiddleware,
		checkArtworkAndUserMiddleware,
		GetArtworkFavoriteStatus)
	r.POST("/favorite",
		middleware.JWTAuthMiddleware.MiddlewareFunc(),
		validateArtworkIDMiddleware,
		checkArtworkAndUserMiddleware,
		FavoriteArtwork)
	r.DELETE("/favorite",
		middleware.JWTAuthMiddleware.MiddlewareFunc(),
		validateArtworkIDMiddleware,
		checkArtworkAndUserMiddleware,
		UnfavoriteArtwork)
}
