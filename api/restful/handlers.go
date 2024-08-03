package restful

import (
	"ManyACG/service"
	"ManyACG/telegram"
	"ManyACG/types"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegoutil"

	. "ManyACG/logger"
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

func SendArtworkInfo(ctx *gin.Context) {
	var request SendArtworkInfoRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(
			http.StatusBadRequest,
			gin.H{
				"message": err.Error(),
			},
		)
		return
	}
	var chatID telego.ChatID
	if request.ChatID != 0 {
		chatID = telegoutil.ID(request.ChatID)
	}
	copyCtx := ctx.Copy()
	go func() {
		if err := telegram.SendArtworkInfo(
			copyCtx,
			telegram.Bot,
			request.SourceURL,
			true,
			&chatID,
			true,
			request.AppendCaption,
			true,
			nil,
		); err != nil {
			Logger.Error(err)
		}
	}()
	ctx.JSON(http.StatusOK, gin.H{
		"message": "task created",
	})
}
