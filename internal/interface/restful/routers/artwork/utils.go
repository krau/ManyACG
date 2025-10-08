package artwork

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/krau/ManyACG/api/restful/utils"
	"github.com/krau/ManyACG/common"
	"github.com/krau/ManyACG/internal/service"
	"github.com/krau/ManyACG/types"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// 检查是否能查看 R18 作品 (仅适用于使用了 OptionalJWTMiddleware 的路由)
//
// 在内部做响应处理，如果不能查看则返回 false
// func checkR18Permission(ctx *gin.Context) bool {
// 	logged := ctx.GetBool("logged")
// 	if !logged {
// 		ctx.JSON(http.StatusUnauthorized, utils.RestfulCommonResponse[any]{
// 			Status:  http.StatusUnauthorized,
// 			Message: "You must log in to view R18 content",
// 		})
// 		return false
// 	}
// 	claims := ctx.MustGet("claims").(jwt.MapClaims)
// 	username := claims["id"].(string)
// 	user, err := service.GetUserByUsername(ctx, username)
// 	if err != nil {
// 		if errors.Is(err, mongo.ErrNoDocuments) {
// 			ctx.JSON(http.StatusForbidden, utils.RestfulCommonResponse[any]{
// 				Status:  http.StatusForbidden,
// 				Message: "Account not found",
// 			})
// 			return false
// 		}
// 		common.Logger.Errorf("Failed to get user: %v", err)
// 		ctx.JSON(http.StatusInternalServerError, utils.RestfulCommonResponse[any]{
// 			Status:  http.StatusInternalServerError,
// 			Message: "Failed to get user",
// 		})
// 		return false
// 	}
// 	if !user.Settings.R18 {
// 		ctx.JSON(http.StatusForbidden, utils.RestfulCommonResponse[any]{
// 			Status:  http.StatusForbidden,
// 			Message: "Your settings do not allow you to view R18 content",
// 		})
// 		return false
// 	}
// 	return true
// }

func validateArtworkIDMiddleware(ctx *gin.Context) {
	var request ArtworkIDRequest
	if err := ctx.ShouldBind(&request); err != nil {
		utils.GinBindError(ctx, err)
		return
	}
	objectID, err := primitive.ObjectIDFromHex(request.ArtworkID)
	if err != nil {
		utils.GinErrorResponse(ctx, err, http.StatusBadRequest, "Invalid artwork ID")
		return
	}
	ctx.Set("artwork_id", objectID)
	ctx.Next()
}

// 在 validateArtworkIDMiddleware 之后调用
func checkArtworkAndUserMiddleware(ctx *gin.Context) {
	artworkID := ctx.MustGet("artwork_id").(primitive.ObjectID)
	_, err := service.GetArtworkByID(ctx, artworkID)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			utils.GinErrorResponse(ctx, err, http.StatusBadRequest, "Artwork not found")
			return
		}
		common.Logger.Errorf("Failed to get artwork: %v", err)
		utils.GinErrorResponse(ctx, err, http.StatusInternalServerError, "Failed to get artwork")
		return
	}

	claims := jwt.ExtractClaims(ctx)
	username := claims["id"].(string)
	user, err := service.GetUserByUsername(ctx, username)
	if errors.Is(err, mongo.ErrNoDocuments) {
		utils.GinErrorResponse(ctx, err, http.StatusForbidden, "Account not found")
		return
	}
	if err != nil {
		common.Logger.Errorf("Failed to get user: %v", err)
		utils.GinErrorResponse(ctx, err, http.StatusInternalServerError, "Failed to get user")
		return
	}
	ctx.Set("user", user)
	ctx.Next()
}

func checkApiKeyFetchArtworkPermission(ctx *gin.Context) {
	apiKey := ctx.MustGet("api_key").(*types.ApiKeyModel)
	if !apiKey.HasPermission(types.ApiKeyPermissionFetchArtwork) {
		utils.GinErrorResponse(ctx, fmt.Errorf("api key does not have permission: %s", types.ApiKeyPermissionFetchArtwork), http.StatusForbidden, "Forbidden")
		ctx.Abort()
		return
	}
	ctx.Next()
}
