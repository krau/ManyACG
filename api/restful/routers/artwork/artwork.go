package artwork

import (
	"ManyACG/api/restful/middleware"
	"net/http"

	"github.com/gin-gonic/gin"
)

func RegisterRouter(r *gin.RouterGroup) {
	r.Match([]string{http.MethodGet, http.MethodPost},
		"/random",
		middleware.OptionalJWTMiddleware,
		RandomArtworks)
	r.Match([]string{http.MethodGet, http.MethodPost},
		"/random/preview",
		middleware.OptionalJWTMiddleware,
		RandomArtworkPreview)
	r.Match([]string{http.MethodGet, http.MethodPost},
		"/list",
		middleware.OptionalJWTMiddleware,
		GetArtworkList)
	r.GET("/count", GetArtworkCount)
	r.GET("/:id",
		middleware.OptionalJWTMiddleware,
		GetArtwork)
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
