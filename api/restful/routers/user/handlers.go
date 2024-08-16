package user

import (
	"ManyACG/service"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func GetUnauthUser(ctx *gin.Context) {
	objectID, ok := ctx.Get("object_id")
	if !ok {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "invalid object id"})
		return
	}
	user, err := service.GetUnauthUserByID(ctx, objectID.(primitive.ObjectID))
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			ctx.JSON(http.StatusNotFound, gin.H{"message": "user not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": "failed to get user"})
		return
	}
	ctx.JSON(http.StatusOK, &UnauthUserResponse{
		ID:         user.ID.Hex(),
		Username:   user.Username,
		TelegramID: user.TelegramID,
	})
}

func GetProfile(ctx *gin.Context) {

}
