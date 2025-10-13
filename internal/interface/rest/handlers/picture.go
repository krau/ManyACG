package handlers

import (
	"github.com/gofiber/fiber/v3"
	"github.com/krau/ManyACG/internal/common/httpclient"
	"github.com/krau/ManyACG/internal/infra/config/runtimecfg"
	"github.com/krau/ManyACG/internal/interface/rest/common"
	"github.com/krau/ManyACG/internal/interface/rest/utils"
	"github.com/krau/ManyACG/internal/service"
	"github.com/krau/ManyACG/internal/shared"
	"github.com/krau/ManyACG/pkg/objectuuid"
)

func HandleGetPictureFileByID(ctx fiber.Ctx) error {
	pictureID := ctx.Params("id")
	if pictureID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "missing picture ID")
	}
	id, err := objectuuid.FromObjectIDHex(pictureID)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid picture ID")
	}
	serv := common.MustGetState[*service.Service](ctx, common.StateKeyService)
	picture, err := serv.GetPictureByID(ctx, id)
	if err != nil {
		return err
	}
	var filePath string
	if detail := picture.StorageInfo.Data().Original; detail != nil {
		file, err := serv.StorageGetFile(ctx, *detail)
		if err != nil {
			return err
		}
		defer file.Close()
		filePath = file.Name()
	} else {
		file, err := httpclient.DownloadWithCache(ctx, picture.Original, nil)
		if err != nil {
			return err
		}
		defer file.Close()
		filePath = file.Name()
	}

	return ctx.SendFile(filePath, fiber.SendFile{Compress: true})
}

func HandleGetRandomPicture(ctx fiber.Ctx) error {
	serv := common.MustGetState[*service.Service](ctx, common.StateKeyService)
	picture, err := serv.RandomPictures(ctx, 1)
	if err != nil {
		return err
	}
	if len(picture) == 0 {
		return fiber.NewError(fiber.StatusNotFound, "no picture found")
	}
	pic := picture[0]
	if pic.StorageInfo.Data() == shared.ZeroStorageInfo || pic.StorageInfo.Data().Regular == nil {
		return ctx.Redirect().To(pic.Thumbnail)
	}
	cfg := common.MustGetState[runtimecfg.RestConfig](ctx, common.StateKeyConfig)
	picUrl := utils.ApplyApiStoragePathRule(*pic.GetStorageInfo().Regular, cfg.StoragePathRules)
	if picUrl == "" {
		return ctx.Redirect().To(pic.Thumbnail)
	}
	return ctx.Redirect().To(picUrl)
}
