package artwork

import (
	"ManyACG/api/restful/utils"
	"ManyACG/service"
	"errors"
	"net/http"

	. "ManyACG/logger"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// 检查是否能查看 R18 作品 (仅适用于使用了 OptionalJWTMiddleware 的路由)
//
// 在内部做响应处理，如果不能查看则返回 false
func checkR18Permission(ctx *gin.Context) bool {
	logged := ctx.GetBool("logged")
	if !logged {
		ctx.JSON(http.StatusUnauthorized, &ArtworkResponse{
			Status:  http.StatusUnauthorized,
			Message: "You need to login to view this artworks",
		})
		return false
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
			return false
		}
		Logger.Errorf("Failed to get user: %v", err)
		ctx.JSON(http.StatusInternalServerError, &ArtworkResponse{
			Status:  http.StatusInternalServerError,
			Message: "Failed to get user",
		})
		return false
	}
	if !user.Settings.R18 {
		ctx.JSON(http.StatusBadRequest, &ArtworkResponse{
			Status:  http.StatusBadRequest,
			Message: "Your settings do not allow you to view this artworks",
		})
		return false
	}
	return true
}

func validateArtworkIDMiddleware(ctx *gin.Context) {
	var request ArtworkIDRequest
	if err := ctx.ShouldBind(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, &ArtworkResponse{
			Status:  http.StatusBadRequest,
			Message: utils.BindError(ctx, err),
		})
		ctx.Abort()
		return
	}
	objectID, err := primitive.ObjectIDFromHex(request.ArtworkID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, &ArtworkResponse{
			Status:  http.StatusBadRequest,
			Message: "Invalid Artwork ID",
		})
		ctx.Abort()
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
			ctx.JSON(http.StatusBadRequest, &ArtworkResponse{
				Status:  http.StatusBadRequest,
				Message: "Artwork not found",
			})
			ctx.Abort()
			return
		}
		Logger.Errorf("Failed to get artwork: %v", err)
		ctx.JSON(http.StatusInternalServerError, &ArtworkResponse{
			Status:  http.StatusInternalServerError,
			Message: "Failed to get artwork",
		})
		ctx.Abort()
		return
	}

	claims := jwt.ExtractClaims(ctx)
	username := claims["id"].(string)
	user, err := service.GetUserByUsername(ctx, username)
	if errors.Is(err, mongo.ErrNoDocuments) {
		ctx.JSON(http.StatusBadRequest, &ArtworkResponse{
			Status:  http.StatusBadRequest,
			Message: "User not found",
		})
		ctx.Abort()
		return
	}
	if err != nil {
		Logger.Errorf("Failed to get user: %v", err)
		ctx.JSON(http.StatusInternalServerError, &ArtworkResponse{
			Status:  http.StatusInternalServerError,
			Message: "Failed to get user",
		})
		ctx.Abort()
		return
	}
	ctx.Set("user", user)
	ctx.Next()
}
