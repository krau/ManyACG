package handlers

import (
	"github.com/gofiber/fiber/v3"
	"github.com/krau/ManyACG/internal/interface/rest/common"
)

type RequestSendArtworkInfoByTelegramBot struct {
	SourceURL     string `json:"source_url" query:"source_url" form:"source_url" validate:"required,url"`
	ChatID        int64  `json:"chat_id" query:"chat_id" form:"chat_id" validate:"required"`
	AppendCaption string `json:"append_caption" query:"append_caption" form:"append_caption"`
}

func HandleSendArtworkInfoByTelegramBot(ctx fiber.Ctx) error {
	bot, ok := common.GetState[common.TelegramBot](ctx, common.StateKeyTelegramBot)
	if !ok {
		return fiber.ErrInternalServerError
	}
	req := new(RequestSendArtworkInfoByTelegramBot)
	if err := ctx.Bind().All(req); err != nil {
		return err
	}
	if err := bot.SendArtworkInfo(ctx, req.SourceURL, req.ChatID, req.AppendCaption); err != nil {
		return err
	}
	return ctx.JSON(common.NewSuccess("ok"))
}
