package artist

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/krau/ManyACG/api/restful/utils"
	"github.com/krau/ManyACG/common"
	"github.com/krau/ManyACG/pkg/objectuuid"

	"github.com/krau/ManyACG/internal/service"
	"github.com/krau/ManyACG/types"
	"go.mongodb.org/mongo-driver/mongo"
)

func GetArtist(ctx *gin.Context) {
	artistID := ctx.MustGet("object_id").(objectuuid.ObjectUUID)
	artist, err := service.GetArtistByID(ctx, artistID)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			utils.GinErrorResponse(ctx, err, http.StatusNotFound, "Artist not found")
			return
		}
		common.Logger.Errorf("Failed to get artist: %v", err)
		utils.GinErrorResponse(ctx, err, http.StatusInternalServerError, "Failed to get artist")
		return
	}
	ctx.JSON(http.StatusOK, utils.RestfulCommonResponse[*types.Artist]{
		Status:  http.StatusOK,
		Message: "Success",
		Data:    artist,
	})
}
