package handlers

import (
	"github.com/gofiber/fiber/v3"
	"github.com/krau/ManyACG/internal/infra/config/runtimecfg"
	"github.com/krau/ManyACG/internal/interface/rest/common"
	"github.com/krau/ManyACG/internal/interface/rest/utils"
	"github.com/krau/ManyACG/internal/model/entity"
	"github.com/krau/ManyACG/internal/model/query"
	"github.com/krau/ManyACG/internal/service"
	"github.com/krau/ManyACG/internal/shared"
	"github.com/krau/ManyACG/internal/shared/errs"
	"github.com/unvgo/ouid"
)

type RequestRandomArtworks struct {
	R18   int `query:"r18,default=0" form:"r18,default=0" json:"r18" validate:"gte=0,lte=2" message:"r18 must be 0 (no R18), 1 (only R18) or 2 (both)"`
	Limit int `query:"limit,default=1" form:"limit,default=1" json:"limit" validate:"lte=200" message:"limit must be between 1 and 200"`
}

type RequestCountArtwork struct {
	R18 int `query:"r18" form:"r18" json:"r18" validate:"omitempty,gte=0,lte=2" message:"r18 must be 0 (no R18), 1 (only R18) or 2 (both)"`
}

func HandleRandomArtworks(ctx fiber.Ctx) error {
	serv := common.MustGetState[*service.Service](ctx, common.StateKeyService)
	req := new(RequestRandomArtworks)
	if err := ctx.Bind().All(req); err != nil {
		return err
	}
	if req.Limit <= 0 {
		req.Limit = 1
	}
	artworks, err := serv.QueryArtworks(ctx, query.ArtworksDB{
		ArtworksFilter: query.ArtworksFilter{
			R18:        shared.R18TypeFromInt(req.R18),
			HasPicture: true,
		},
		Random: true,
		Paginate: query.Paginate{
			Limit: req.Limit,
		},
	})
	if err != nil {
		return err
	}
	cfg := common.MustGetState[runtimecfg.RestConfig](ctx, common.StateKeyConfig)
	resp := artworksResponseFromEntity(ctx, artworks, cfg, serv)
	if len(resp) == 0 {
		return common.NewError(fiber.StatusNotFound, "no artworks found")
	}
	return ctx.JSON(common.NewSuccess(resp))
}

func HandleRandomPreviewArtworks(ctx fiber.Ctx) error {
	serv := common.MustGetState[*service.Service](ctx, common.StateKeyService)
	req := new(RequestRandomArtworks)
	if err := ctx.Bind().All(req); err != nil {
		return err
	}
	artworks, err := serv.QueryArtworks(ctx, query.ArtworksDB{
		ArtworksFilter: query.ArtworksFilter{
			R18:        shared.R18TypeFromInt(req.R18),
			HasPicture: true,
		},
		Random: true,
		Paginate: query.Paginate{
			Limit: 1,
		},
	})
	if err != nil {
		return err
	}
	if len(artworks) == 0 {
		return common.NewError(fiber.StatusNotFound, "no artworks found")
	}
	cfg := common.MustGetState[runtimecfg.RestConfig](ctx, common.StateKeyConfig)
	if artworks[0] == nil || len(artworks[0].Pictures) == 0 {
		return common.NewError(fiber.StatusNotFound, "no pictures found for artwork")
	}
	pic := artworks[0].Pictures[0]
	_, regular := utils.PictureResponseUrl(ctx, pic, cfg)
	return ctx.Redirect().To(regular)
}

func HandleCountArtwork(ctx fiber.Ctx) error {
	serv := common.MustGetState[*service.Service](ctx, common.StateKeyService)
	req := new(RequestCountArtwork)
	if err := ctx.Bind().All(req); err != nil {
		return err
	}
	count, err := serv.CountArtworks(ctx, shared.R18TypeFromInt(req.R18))
	if err != nil {
		return err
	}
	return ctx.JSON(common.NewSuccess(count))
}

func HandleGetArtworkByID(ctx fiber.Ctx) error {
	artworkID := ctx.Params("id")
	if artworkID == "" {
		return common.NewError(fiber.StatusBadRequest, "artwork id is required")
	}
	artworkUUID, err := ouid.FromObjectIDHex(artworkID)
	if err != nil {
		return common.NewError(fiber.StatusBadRequest, "invalid artwork id")
	}
	serv := common.MustGetState[*service.Service](ctx, common.StateKeyService)
	artwork, err := serv.GetArtworkByID(ctx, artworkUUID)
	if err != nil {
		if err == errs.ErrRecordNotFound {
			return common.NewError(fiber.StatusNotFound, "artwork not found")
		}
		return err
	}
	cfg := common.MustGetState[runtimecfg.RestConfig](ctx, common.StateKeyConfig)
	resp := artworksResponseFromEntity(ctx, []*entity.Artwork{artwork}, cfg, serv)
	return ctx.JSON(common.NewSuccess(resp[0]))
}
