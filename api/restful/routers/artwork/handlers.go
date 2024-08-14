package artwork

import (
	. "ManyACG/logger"
	"ManyACG/service"
	"ManyACG/types"
	"errors"
	"net/http"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func RandomArtworks(ctx *gin.Context) {
	var request GetRandomArtworksRequest
	if err := ctx.ShouldBind(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status":  http.StatusBadRequest,
			"message": err.Error(),
		})
		return
	}
	r18Type := types.R18Type(request.R18)
	artwork, err := service.GetRandomArtworks(ctx, r18Type, request.Limit)
	isAuthorized := ctx.GetBool("auth")
	if err != nil {
		if isAuthorized {
			ctx.JSON(http.StatusInternalServerError, &ArtworkResponse{
				Status:  http.StatusInternalServerError,
				Message: err.Error(),
			})
			return
		} else {
			ctx.JSON(http.StatusInternalServerError, &ArtworkResponse{
				Status:  http.StatusInternalServerError,
				Message: "Failed to get artwork",
			})
			return
		}
	}
	if len(artwork) == 0 {
		ctx.JSON(http.StatusNotFound, &ArtworkResponse{
			Status:  http.StatusNotFound,
			Message: "Artwork not found",
		})
		return
	}
	ctx.JSON(http.StatusOK, ResponseFromArtworks(artwork, isAuthorized))
}

func GetArtwork(ctx *gin.Context) {
	id := ctx.Param("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, &ArtworkResponse{
			Status:  http.StatusBadRequest,
			Message: "Invalid ID",
		})
		return
	}

	artwork, err := service.GetArtworkByID(ctx, objectID)
	isAuthorized := ctx.GetBool("auth")
	if err != nil {
		if isAuthorized {
			ctx.JSON(http.StatusInternalServerError, &ArtworkResponse{
				Status:  http.StatusInternalServerError,
				Message: err.Error(),
			})
			return
		} else {
			ctx.JSON(http.StatusInternalServerError, &ArtworkResponse{
				Status:  http.StatusInternalServerError,
				Message: "Failed to get artwork",
			})
			return
		}
	}
	if artwork == nil {
		ctx.JSON(http.StatusNotFound, &ArtworkResponse{
			Status:  http.StatusNotFound,
			Message: "Artwork not found",
		})
		return
	}
	ctx.JSON(http.StatusOK, ResponseFromArtwork(artwork, isAuthorized))
}

func GetLatestArtworks(ctx *gin.Context) {
	var request GetLatestArtworksRequest
	if err := ctx.ShouldBind(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status":  http.StatusBadRequest,
			"message": err.Error(),
		})
		return
	}

	hasKey := ctx.GetBool("auth")
	r18Type := types.R18Type(request.R18)
	if r18Type != types.R18TypeNone && !hasKey {
		logged := ctx.GetBool("logged")
		if !logged {
			ctx.JSON(http.StatusUnauthorized, &ArtworkResponse{
				Status:  http.StatusUnauthorized,
				Message: "You need to login to view R18 artworks",
			})
			return
		}
		claims := ctx.MustGet("claims").(jwt.MapClaims)
		username := claims["id"].(string)
		user, err := service.GetUserByUsername(ctx, username)
		if err != nil {
			if errors.Is(err, mongo.ErrNoDocuments) {
				ctx.JSON(http.StatusForbidden, &ArtworkResponse{
					Status:  http.StatusForbidden,
					Message: "Account not found",
				})
				return
			}
			Logger.Errorf("Failed to get user: %v", err)
			ctx.JSON(http.StatusInternalServerError, &ArtworkResponse{
				Status:  http.StatusInternalServerError,
				Message: "Failed to get user",
			})
		}
		if !user.Settings.R18 {
			ctx.JSON(http.StatusBadRequest, &ArtworkResponse{
				Status:  http.StatusBadRequest,
				Message: "Your settings do not allow you to view R18 artworks",
			})
			return
		}
	}

	artworks, err := service.GetLatestArtworks(ctx, r18Type, request.Page, request.PageSize)
	if err != nil {
		if hasKey {
			ctx.JSON(http.StatusInternalServerError, &ArtworkResponse{
				Status:  http.StatusInternalServerError,
				Message: err.Error(),
			})
			return
		} else {
			ctx.JSON(http.StatusInternalServerError, &ArtworkResponse{
				Status:  http.StatusInternalServerError,
				Message: "Failed to get artworks",
			})
			return
		}
	}
	if len(artworks) == 0 {
		ctx.JSON(http.StatusNotFound, &ArtworkResponse{
			Status:  http.StatusNotFound,
			Message: "Artworks not found",
		})
		return
	}
	ctx.JSON(http.StatusOK, ResponseFromArtworks(artworks, hasKey))
}
