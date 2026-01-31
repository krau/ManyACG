package handlers

import (
	"context"

	"github.com/gofiber/fiber/v3"
	"github.com/krau/ManyACG/internal/interface/rest/common"
	"github.com/krau/ManyACG/internal/service"
	"github.com/krau/ManyACG/internal/shared"
)

type RequestSendArtworkInfoByTelegramBot struct {
	SourceURL     string `json:"source_url" query:"source_url" form:"source_url" validate:"required"`
	ChatID        int64  `json:"chat_id" query:"chat_id" form:"chat_id" validate:"required"`
	AppendCaption string `json:"append_caption" query:"append_caption" form:"append_caption"`
}

func HandleSendArtworkInfoByTelegramBot(ctx fiber.Ctx) error {
	key := ctx.Get("X-API-KEY")
	if key == "" {
		return common.NewError(fiber.StatusUnauthorized, "api key is required")
	}
	serv := common.MustGetState[*service.Service](ctx, common.StateKeyService)
	keyEnt, err := serv.GetApiKeyByKey(ctx, key)
	if err != nil {
		return common.NewError(fiber.StatusUnauthorized, "invalid api key")
	}
	if !keyEnt.HasPermission(shared.PermissionSendArtworkInfo) {
		return common.NewError(fiber.StatusForbidden, "api key does not have permission")
	}
	if !keyEnt.CanUse() {
		return common.NewError(fiber.StatusForbidden, "api key quota exceeded")
	}
	bot, ok := common.GetState[common.TelegramBot](ctx, common.StateKeyTelegramBot)
	if !ok {
		return fiber.ErrInternalServerError
	}
	req := new(RequestSendArtworkInfoByTelegramBot)
	if err := ctx.Bind().All(req); err != nil {
		return err
	}
	serv.IncreaseApiKeyUsed(ctx, key)
	// current implement of SendArtworkInfo use a buffered channel, so it will return immediately and run in the background.
	// thus we should use context.Background() here.
	bot.SendArtworkInfo(context.Background(), req.SourceURL, req.ChatID, req.AppendCaption)
	return ctx.JSON(common.NewSuccess("ok"))
}
