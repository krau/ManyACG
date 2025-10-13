package handlers

import (
	"github.com/gofiber/fiber/v3"
	"github.com/krau/ManyACG/internal/interface/rest/common"
	"github.com/krau/ManyACG/internal/service"
	"github.com/krau/ManyACG/pkg/objectuuid"
)

func HandleGetArtistByID(ctx fiber.Ctx) error {
	artistID := ctx.Params("id")
	if artistID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "missing artist ID")
	}
	id, err := objectuuid.FromObjectIDHex(artistID)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid artist ID")
	}
	serv := common.MustGetState[*service.Service](ctx, common.StateKeyService)
	artist, err := serv.GetArtistByID(ctx, id)
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
