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
		"/list",
		middleware.OptionalJWTMiddleware,
		GetLatestArtworks)
	r.POST("/like",
		middleware.JWTAuthMiddleware.MiddlewareFunc(),
		validateArtworkIDMiddleware,
		checkArtworkAndUserMiddleware,
		LikeArtwork)
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
	r.GET("/:id",
		middleware.OptionalJWTMiddleware,
		GetArtwork)
}
