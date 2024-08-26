package artist

import (
	"ManyACG/common"
	"ManyACG/service"
	"ManyACG/types"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func GetArtist(ctx *gin.Context) {
	artistID := ctx.MustGet("object_id").(primitive.ObjectID)
	artist, err := service.GetArtistByID(ctx, artistID)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			common.GinErrorResponse(ctx, err, http.StatusNotFound, "Artist not found")
			return
		}
		common.GinErrorResponse(ctx, err, http.StatusInternalServerError, "Failed to get artist")
		return
	}
	ctx.JSON(http.StatusOK, common.RestfulCommonResponse[*types.Artist]{
		Status:  http.StatusOK,
		Message: "Success",
		Data:    artist,
	})
}

func GetArtistArtworkCount(ctx *gin.Context) {
	artistID := ctx.MustGet("object_id").(primitive.ObjectID)
	artworkCount, err := service.GetArtistArtworkCount(ctx, artistID)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			common.GinErrorResponse(ctx, err, http.StatusNotFound, "Artist not found")
			return
		}
		common.GinErrorResponse(ctx, err, http.StatusInternalServerError, "Failed to get artist artwork count")
		return
	}
	ctx.JSON(http.StatusOK, common.RestfulCommonResponse[int64]{
		Status:  http.StatusOK,
		Message: "Success",
		Data:    artworkCount,
	})
}
