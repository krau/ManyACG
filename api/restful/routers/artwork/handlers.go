package artwork

import (
	"ManyACG/service"
	"ManyACG/types"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
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
	if err := ctx.ShouldBindQuery(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status":  http.StatusBadRequest,
			"message": err.Error(),
		})
		return
	}
	r18Type := types.R18Type(request.R18)
	artworks, err := service.GetLatestArtworks(ctx, r18Type, request.Page, request.PageSize)
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
	ctx.JSON(http.StatusOK, ResponseFromArtworks(artworks, isAuthorized))
}
