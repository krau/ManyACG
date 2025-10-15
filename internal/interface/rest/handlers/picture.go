package handlers

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"io"
	"slices"

	"github.com/gabriel-vasile/mimetype"
	"github.com/gofiber/fiber/v3"
	"github.com/krau/ManyACG/internal/common/httpclient"
	"github.com/krau/ManyACG/internal/infra/config/runtimecfg"
	"github.com/krau/ManyACG/internal/interface/rest/common"
	"github.com/krau/ManyACG/internal/interface/rest/utils"
	"github.com/krau/ManyACG/internal/service"
	"github.com/krau/ManyACG/internal/shared"
	"github.com/krau/ManyACG/pkg/log"
	"github.com/krau/ManyACG/pkg/objectuuid"
	"gorm.io/datatypes"
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
	ctx.Set(fiber.HeaderContentDisposition, "inline; filename=\""+serv.PrettyFileName(picture.Artwork, picture)+"\"")
	return ctx.SendFile(filePath, fiber.SendFile{Compress: true})
}

func HandleGetSizedPictureFileByID(ctx fiber.Ctx) error {
	size := ctx.Params("size")
	if !slices.Contains([]string{"thumb", "regular", "original"}, size) {
		size = "regular"
	}
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
	var detail *shared.StorageDetail
	switch size {
	case "thumb":
		detail = picture.StorageInfo.Data().Thumb
	case "regular":
		detail = picture.StorageInfo.Data().Regular
	case "original":
		detail = picture.StorageInfo.Data().Original
	}
	if detail != nil {
		if detail.Mime != "" {
			ctx.Set(fiber.HeaderContentType, detail.Mime)
			var sendErr error
			sendWriter := func(w *bufio.Writer) {
				defer w.Flush()
				sendErr = serv.StorageStreamFile(ctx, *detail, w)
			}
			ctx.SendStreamWriter(sendWriter)
			return sendErr
		}

		pr, pw := io.Pipe()
		streamCtx, cancel := context.WithCancel(ctx.Context())
		defer cancel()

		errChan := make(chan error, 1)
		go func() {
			defer close(errChan)
			err := serv.StorageStreamFile(streamCtx, *detail, pw)
			if err != nil {
				pw.CloseWithError(err)
				errChan <- err
				return
			}
			pw.Close()
		}()

		buf := make([]byte, 3072)
		n, err := io.ReadAtLeast(pr, buf, 3072)
		if err != nil {
			cancel()
			pr.Close()

			streamErr := <-errChan
			if streamErr != nil {
				return streamErr
			}

			if !errors.Is(err, io.EOF) && !errors.Is(err, io.ErrUnexpectedEOF) {
				return fiber.NewError(fiber.StatusInternalServerError, err.Error())
			}
		}

		buf = buf[:n]
		mtype := mimetype.Detect(buf)
		if mtype != nil {
			ctx.Set(fiber.HeaderContentType, mtype.String())
			go func() {
				// 异步更新 mime 类型
				detail.Mime = mtype.String()
				storData := picture.StorageInfo.Data()
				if size == "thumb" {
					storData.Thumb = detail
				} else if size == "regular" {
					storData.Regular = detail
				} else if size == "original" {
					storData.Original = detail
				}
				picture.StorageInfo = datatypes.NewJSONType(storData)
				if err := serv.SavePicture(context.Background(), picture); err != nil {
					log.Errorf("failed to save picture mime: %v", err)
				}
			}()
		}
		ctx.Set(fiber.HeaderContentDisposition, "inline")

		fullReader := io.MultiReader(bytes.NewReader(buf), pr)
		return ctx.SendStream(fullReader)
	}
	return ctx.Redirect().To(picture.Thumbnail)
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
	picUrl := utils.ResponseUrlForStoragePath(ctx, *pic.GetStorageInfo().Regular, cfg.StoragePathRules)
	if picUrl == "" {
		return ctx.Redirect().To(pic.Thumbnail)
	}
	return ctx.Redirect().To(picUrl)
}
