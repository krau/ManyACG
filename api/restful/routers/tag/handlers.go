package tag

import (
	. "ManyACG/logger"
	"ManyACG/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type GetRandomTagsRequest struct {
	Limit int `form:"limit,default=20" binding:"gte=1,lte=200" json:"limit"`
}

func GetRandomTags(ctx *gin.Context) {
	var request GetRandomTagsRequest
	if err := ctx.ShouldBind(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status":  http.StatusBadRequest,
			"message": "Invalid request",
		})
		return
	}
	tags, err := service.GetRandomTagModels(ctx, request.Limit)
	if err != nil {
		Logger.Errorf("Failed to get tags: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status":  http.StatusInternalServerError,
			"message": "Failed to get tags",
		})
		return
	}
	if len(tags) == 0 {
		ctx.JSON(http.StatusNotFound, gin.H{
			"status":  http.StatusNotFound,
			"message": "Tags not found",
		})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"status":  http.StatusOK,
		"message": "Success",
		"tags":    tags,
	})
}
