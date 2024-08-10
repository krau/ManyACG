package artwork

import (
	"ManyACG/service"
	"ManyACG/types"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func RandomArtwork(ctx *gin.Context) {
	var r18Str string
	if ctx.Request.Method == http.MethodGet {
		r18Str = ctx.DefaultQuery("r18", "0")
	}
	if ctx.Request.Method == http.MethodPost {
		r18Str = ctx.DefaultPostForm("r18", "0")
	}
	r18Type := types.R18TypeNone
	switch r18Str {
	case "0":
		r18Type = types.R18TypeNone
	case "1":
		r18Type = types.R18TypeOnly
	case "2":
		r18Type = types.R18TypeAll
	}

	artwork, err := service.GetRandomArtworks(ctx, r18Type, 1)
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
	ctx.JSON(http.StatusOK, ResponseFromArtwork(artwork[0], isAuthorized))
}

func GetArtwork(ctx *gin.Context) {
	id := ctx.Param("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid ID",
		})
		return
	}
	artwork, err := service.GetArtworkByID(ctx, objectID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}
	if artwork == nil {
		ctx.JSON(http.StatusNotFound, gin.H{
			"message": "Artwork not found",
		})
		return
	}
	ctx.JSON(http.StatusOK, artwork)
}
