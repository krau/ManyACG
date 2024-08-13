package picture

import (
	. "ManyACG/logger"
	"ManyACG/service"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func ValidatePictureID(ctx *gin.Context) {
	pictureID := ctx.Param("id")
	objectID, err := primitive.ObjectIDFromHex(pictureID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status":  http.StatusBadRequest,
			"message": "Invalid ID",
		})
		ctx.Abort()
		return
	}
	picture, err := service.GetPictureByID(ctx, objectID)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			ctx.JSON(http.StatusNotFound, gin.H{
				"status":  http.StatusNotFound,
				"message": "Picture not found",
			})
		} else {
			Logger.Errorf("Failed to get picture: %v", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"status":  http.StatusInternalServerError,
				"message": "Failed to get picture",
			})
		}
		ctx.Abort()
		return
	}
	ctx.Set("picture", picture)
	ctx.Next()

}