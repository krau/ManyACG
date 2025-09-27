package apikey

import (
	"errors"
	"net/http"

	"github.com/duke-git/lancet/v2/slice"
	"github.com/gin-gonic/gin"
	"github.com/krau/ManyACG/api/restful/middleware"
	"github.com/krau/ManyACG/api/restful/utils"
	"github.com/krau/ManyACG/service"
	"github.com/krau/ManyACG/types"
)

func RegisterRouter(r *gin.RouterGroup) {
	r.POST("/", middleware.AdminKeyRequired, CreateApiKey)
}

type CreateApiKeyRequest struct {
	Key         string   `form:"key" json:"key" binding:"required,min=10,max=100"`
	Quota       int      `form:"quota" json:"quota" binding:"required"`
	Permissions []string `form:"permissions" json:"permissions" binding:"required"`
	Description string   `form:"description" json:"description"`
}

func CreateApiKey(ctx *gin.Context) {
	var request CreateApiKeyRequest
	if err := ctx.ShouldBind(&request); err != nil {
		utils.GinBindError(ctx, err)
		return
	}

	if !slice.ContainSubSlice(types.ValidApiKeyPermissions, request.Permissions) {
		utils.GinBindError(ctx, errors.New("invalid permission"))
		return
	}
	permissions := []types.ApiKeyPermission{}
	for _, permission := range request.Permissions {
		permissions = append(permissions, types.ApiKeyPermission(permission))
	}
	apiKey, err := service.CreateApiKey(ctx, request.Key, request.Quota, permissions, request.Description)
	if err != nil {
		utils.GinErrorResponse(ctx, err, http.StatusInternalServerError, "Failed to create API key")
		return
	}
	ctx.JSON(http.StatusOK, apiKey)
}
