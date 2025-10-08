package user

import (
	"errors"
	"net/http"

	"github.com/krau/ManyACG/api/restful/utils"
	"github.com/krau/ManyACG/common"
	"github.com/krau/ManyACG/types"

	"github.com/krau/ManyACG/internal/service"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func GetUnauthUser(ctx *gin.Context) {
	objectID, ok := ctx.Get("object_id")
	if !ok {
		ctx.JSON(http.StatusBadRequest, utils.RestfulCommonResponse[any]{Status: http.StatusBadRequest, Message: "invalid object id"})
		return
	}
	user, err := service.GetUnauthUserByID(ctx, objectID.(primitive.ObjectID))
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			utils.GinErrorResponse(ctx, err, http.StatusNotFound, "user not found")
			return
		}
		utils.GinErrorResponse(ctx, err, http.StatusInternalServerError, "failed to get user")
		return
	}
	ctx.JSON(http.StatusOK, &UnauthUserResponse{
		ID:         user.ID.Hex(),
		Username:   user.Username,
		TelegramID: user.TelegramID,
	})
}

func GetProfile(ctx *gin.Context) {
	claims := jwt.ExtractClaims(ctx)
	username := claims["id"].(string)
	user, err := service.GetUserByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			ctx.JSON(http.StatusNotFound, gin.H{"message": "user not found"})
			return
		}
		common.Logger.Errorf("failed to get user: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": "failed to get user"})
		return
	}
	ctx.JSON(http.StatusOK, utils.RestfulCommonResponse[*UserResponseData]{Status: http.StatusOK, Message: "success", Data: &UserResponseData{
		Username:   user.Username,
		Email:      user.Email,
		TelegramID: user.TelegramID,
		Settings:   user.Settings,
	}})

}

func UpdateSettings(ctx *gin.Context) {
	var settings UserSettingsRequest
	if err := ctx.ShouldBind(&settings); err != nil {
		common.Logger.Errorf("failed to bind json: %v", err)
		utils.GinBindError(ctx, err)
		return
	}

	claims := jwt.ExtractClaims(ctx)
	username := claims["id"].(string)

	user, err := service.GetUserByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			utils.GinErrorResponse(ctx, err, http.StatusNotFound, "user not found")
			return
		}
		common.Logger.Errorf("failed to get user: %v", err)
		utils.GinErrorResponse(ctx, err, http.StatusInternalServerError, "failed to get user")
		return
	}
	res, err := service.UpdateUserSettings(ctx, user.ID, (*types.UserSettings)(&settings))
	if err != nil {
		common.Logger.Errorf("failed to update user settings: %v", err)
		utils.GinErrorResponse(ctx, err, http.StatusInternalServerError, "failed to update user settings")
		return
	}
	ctx.JSON(http.StatusOK, utils.RestfulCommonResponse[*types.UserSettings]{Status: http.StatusOK, Message: "success", Data: res})
}
