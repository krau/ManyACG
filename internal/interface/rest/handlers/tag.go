package handlers

import (
	"github.com/gofiber/fiber/v3"
	"github.com/krau/ManyACG/internal/interface/rest/common"
	"github.com/krau/ManyACG/internal/service"
)

type RequestRandomTags struct {
	Limit int `json:"limit" query:"limit" form:"limit" validate:"omitempty,min=1,max=200" message:"limit must be between 1 and 200"`
}

type ResponseTag struct {
	ID    string   `json:"id"`
	Name  string   `json:"name"`
	Alias []string `json:"alias"`
}

func HandleGetRandomTags(ctx fiber.Ctx) error {
	var req RequestRandomTags
	if err := ctx.Bind().All(req); err != nil {
		return err
	}
	serv := common.MustGetState[*service.Service](ctx, common.StateKeyService)
	tags, err := serv.RandomTags(ctx, req.Limit)
	if err != nil {
		return err
	}
	var res []ResponseTag
	for _, tag := range tags {
		var alias []string
		for _, a := range tag.Alias {
			alias = append(alias, a.Alias)
		}
		res = append(res, ResponseTag{
			ID:    tag.ID.Hex(),
			Name:  tag.Name,
			Alias: alias,
		})
	}
	return ctx.JSON(common.NewSuccess(res))
}
