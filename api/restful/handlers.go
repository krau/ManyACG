package restful

import (
	"ManyACG/service"
	"ManyACG/types"
	"net/http"

	"github.com/gin-gonic/gin"
)

func Ping(ctx *gin.Context) {
	ctx.JSON(200, gin.H{
		"message": "pong",
	})
}

func RandomArtwork(ctx *gin.Context) {
	var r18Str string
	if ctx.Request.Method == http.MethodGet {
		r18Str = ctx.DefaultQuery("r18", "2")
	}
	if ctx.Request.Method == http.MethodPost {
		r18Str = ctx.DefaultPostForm("r18", "2")
	}
	r18Type := types.R18TypeNone
	switch r18Str {
	case "0":
		r18Type = types.R18TypeNone
	case "1":
		r18Type = types.R18TypeOnly
	case "2":
		r18Type = types.R18TypeAll
	}
	artwork, err := service.GetRandomArtworks(ctx, r18Type, 1)
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
