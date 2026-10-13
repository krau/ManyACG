package handlers

import (
	"github.com/gofiber/fiber/v3"
	"github.com/krau/ManyACG/internal/interface/rest/common"
	"github.com/unvgo/ouid"
)

func HandleGetArtistByID(ctx fiber.Ctx) error {
	requestCtx := ctx.RequestCtx()
	artistID := ctx.Params("id")
	if artistID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "missing artist ID")
	}
	id, err := ouid.FromObjectIDHex(artistID)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid artist ID")
	}
	serv := common.MustGetState[common.QueryService](ctx, common.StateKeyQueryService)
	artist, err := serv.GetArtistByID(requestCtx, id)
	if err != nil {
		return err
	}
	return ctx.JSON(common.NewSuccess(&ResponseArtist{
		ID:       artist.ID.Hex(),
		UID:      artist.UID,
		Type:     artist.Type,
		Username: artist.Username,
		Name:     artist.Name,
	}))
}
