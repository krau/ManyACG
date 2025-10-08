package tag

import (
	"net/http"

	"github.com/krau/ManyACG/api/restful/utils"
	"github.com/krau/ManyACG/common"
	"github.com/krau/ManyACG/types"

	"github.com/krau/ManyACG/internal/service"

	"github.com/gin-gonic/gin"
)

type GetRandomTagsRequest struct {
	Limit int `form:"limit,default=20" binding:"gte=1,lte=200" json:"limit"`
}

func GetRandomTags(ctx *gin.Context) {
	var request GetRandomTagsRequest
	if err := ctx.ShouldBind(&request); err != nil {
		utils.GinBindError(ctx, err)
		return
	}
	tags, err := service.GetRandomTagModels(ctx, request.Limit)
	if err != nil {
		common.Logger.Errorf("Failed to get tags: %v", err)
		utils.GinErrorResponse(ctx, err, http.StatusInternalServerError, "Failed to get random tags")
		return
	}
	if len(tags) == 0 {
		ctx.JSON(http.StatusNotFound, utils.RestfulCommonResponse[any]{Status: http.StatusNotFound, Message: "Tags not found"})
		return
	}
	ctx.JSON(http.StatusOK, utils.RestfulCommonResponse[[]*types.TagModel]{Status: http.StatusOK, Message: "Success", Data: tags})
}
