package bot

import (
	"net/http"

	"github.com/krau/ManyACG/api/restful/utils"
	"github.com/krau/ManyACG/internal/common"
	"github.com/krau/ManyACG/internal/infra/config"
	tgUtils "github.com/krau/ManyACG/internal/intf/telegram/utils"
	"github.com/krau/ManyACG/telegram"

	"github.com/gin-gonic/gin"
	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegoutil"
)

func SendArtworkInfo(ctx *gin.Context) {
	if config.Get().Telegram.Token == "" {
		ctx.JSON(http.StatusServiceUnavailable, utils.RestfulCommonResponse[any]{Status: http.StatusServiceUnavailable, Message: "Telegram bot is not available"})
		return
	}
	var request SendArtworkInfoRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		utils.GinBindError(ctx, err)
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
	ctx.JSON(http.StatusOK, utils.RestfulCommonResponse[any]{Status: http.StatusOK, Message: "Task created"})
}
