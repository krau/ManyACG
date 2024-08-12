package picture

import (
	"ManyACG/service"
	"errors"
	"net/http"

	. "ManyACG/logger"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func GetThumb(ctx *gin.Context) {
	pictureID := ctx.Param("id")
	objectID, err := primitive.ObjectIDFromHex(pictureID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status":  http.StatusBadRequest,
			"message": "Invalid ID",
		})
		return
	}
	picture, err := service.GetPictureByID(ctx, objectID)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			ctx.JSON(http.StatusNotFound, gin.H{
				"status":  http.StatusNotFound,
				"message": "Picture not found",
			})
			return
		} else {
			Logger.Errorf("Failed to get picture: %v", err)
			Logger.Errorf("Failed to get picture: %v", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"status":  http.StatusInternalServerError,
				"message": "Failed to get picture",
			})
			return
		}
	}
	ctx.Redirect(http.StatusFound, picture.Thumbnail)
}
