package bot

import (
	"ManyACG/telegram"
	tgUtils "ManyACG/telegram/utils"
	"net/http"

	. "ManyACG/logger"

	"github.com/gin-gonic/gin"
	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegoutil"
)

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
		if err := telegram.SendArtworkInfo(copyCtx, nil, &tgUtils.SendArtworkInfoParams{
			ChatID:        &chatID,
			SourceURL:     request.SourceURL,
			AppendCaption: request.AppendCaption,
			Verify:        true,
			HasPermission: true,
			IgnoreDeleted: true,
			ReplyParams:   nil,
		}); err != nil {
			Logger.Error(err)
		}
	}()
	ctx.JSON(http.StatusOK, gin.H{
		"message": "task created",
	})
}
