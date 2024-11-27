package tag

import (
	"net/http"

	"github.com/krau/ManyACG/common"
	"github.com/krau/ManyACG/types"

	"github.com/krau/ManyACG/service"

	"github.com/gin-gonic/gin"
)

type GetRandomTagsRequest struct {
	Limit int `form:"limit,default=20" binding:"gte=1,lte=200" json:"limit"`
}

func GetRandomTags(ctx *gin.Context) {
	var request GetRandomTagsRequest
	if err := ctx.ShouldBind(&request); err != nil {
		common.GinBindError(ctx, err)
		return
	}
	tags, err := service.GetRandomTagModels(ctx, request.Limit)
	if err != nil {
		common.Logger.Errorf("Failed to get tags: %v", err)
		common.GinErrorResponse(ctx, err, http.StatusInternalServerError, "Failed to get random tags")
		return
	}
	if len(tags) == 0 {
		ctx.JSON(http.StatusNotFound, common.RestfulCommonResponse[any]{Status: http.StatusNotFound, Message: "Tags not found"})
		return
	}
	ctx.JSON(http.StatusOK, common.RestfulCommonResponse[[]*types.TagModel]{Status: http.StatusOK, Message: "Success", Data: tags})
}
