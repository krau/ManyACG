package bot

import (
	"net/http"

	"github.com/krau/ManyACG/common"
	"github.com/krau/ManyACG/config"
	"github.com/krau/ManyACG/telegram"
	tgUtils "github.com/krau/ManyACG/telegram/utils"

	"github.com/gin-gonic/gin"
	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegoutil"
)

func SendArtworkInfo(ctx *gin.Context) {
	if config.Cfg.Telegram.Token == "" {
		ctx.JSON(http.StatusServiceUnavailable, common.RestfulCommonResponse[any]{Status: http.StatusServiceUnavailable, Message: "Telegram bot is not available"})
		return
	}
	var request SendArtworkInfoRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		common.GinBindError(ctx, err)
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
			common.Logger.Error(err)
		}
	}()
	ctx.JSON(http.StatusOK, common.RestfulCommonResponse[any]{Status: http.StatusOK, Message: "Task created"})
}
