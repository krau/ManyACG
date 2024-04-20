package restful

import (
	"ManyACG-Bot/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func Ping(ctx *gin.Context) {
	ctx.JSON(200, gin.H{
		"message": "pong",
	})
}

func RandomArtwork(ctx *gin.Context) {
	var r18Str string
	var tags []string
	if ctx.Request.Method == http.MethodGet {
		r18Str = ctx.DefaultQuery("r18", "false")
		tags = ctx.QueryArray("tags")
	}
	if ctx.Request.Method == http.MethodPost {
		r18Str = ctx.DefaultPostForm("r18", "false")
		tags = ctx.PostFormArray("tags")
	}
	r18, err := strconv.ParseBool(r18Str)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid r18 value",
		})
		return
	}
	artwork, err := service.GetRandomArtworksByTagsR18(ctx, tags, r18, 1)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}
	if len(artwork) == 0 {
		ctx.JSON(http.StatusNotFound, gin.H{
			"message": "No artwork found",
		})
		return
	}
	ctx.JSON(http.StatusOK, artwork[0])
}
