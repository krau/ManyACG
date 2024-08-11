package tag

import (
	. "ManyACG/logger"
	"ManyACG/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func GetRandomTags(ctx *gin.Context) {
	limitStr := ctx.DefaultQuery("limit", "20")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status":  http.StatusBadRequest,
			"message": "Invalid limit value",
		})
		return
	}
	tags, err := service.GetRandomTagModels(ctx, limit)
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
