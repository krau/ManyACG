package handlers

import (
	"bytes"
	"context"
	"errors"
	"io"
	"slices"

	"github.com/gabriel-vasile/mimetype"
	"github.com/gofiber/fiber/v3"
	"github.com/krau/ManyACG/internal/infra/config/runtimecfg"
	"github.com/krau/ManyACG/internal/interface/rest/common"
	"github.com/krau/ManyACG/internal/interface/rest/utils"
	"github.com/krau/ManyACG/internal/service"
	"github.com/krau/ManyACG/internal/shared"
	"github.com/krau/ManyACG/internal/shared/errs"
	"github.com/krau/ManyACG/pkg/log"
	"github.com/unvgo/ouid"
	"gorm.io/datatypes"
)

func HandleGetPictureFileByID(ctx fiber.Ctx) error {
	requestCtx := ctx.RequestCtx()
	safeCtx := ctx.Context()
	pictureID := ctx.Params("id")
	if pictureID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "missing picture ID")
	}
	id, err := ouid.FromObjectIDHex(pictureID)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid picture ID")
	}
	serv := common.MustGetState[*service.Service](ctx, common.StateKeyService)
	picture, err := serv.GetPictureByID(requestCtx, id)
	if err != nil {
		return err
	}
	var filePath string
	if detail := picture.StorageInfo.Data().Original; detail != nil {
		file, err := serv.StorageGetFile(requestCtx, *detail)
		if err != nil {
			return err
		}
		defer file.Close()
		filePath = file.Name()
	} else {
		file, err := serv.DownloadOriginFile(safeCtx, picture.Original)
		if err != nil {
			return err
		}
		defer file.Close()
		filePath = file.Name()
	}
	ctx.Set(fiber.HeaderContentDisposition, "inline; filename=\""+serv.PrettyFileName(picture.Artwork, picture)+"\"")
	return ctx.SendFile(filePath, fiber.SendFile{Compress: true})
}

type pictureStream struct {
	io.Reader
	pipe   *io.PipeReader
	cancel context.CancelFunc
}

func (s *pictureStream) Close() error {
	s.cancel()
	return s.pipe.Close()
}

func NewSizedPictureFileHandler(updateMIME bool) fiber.Handler {
	return func(ctx fiber.Ctx) error {
		requestCtx := ctx.RequestCtx()
		size := ctx.Params("size")
		if !slices.Contains([]string{"thumb", "regular", "original"}, size) {
			size = "regular"
		}
		pictureID := ctx.Params("id")
		if pictureID == "" {
			return fiber.NewError(fiber.StatusBadRequest, "missing picture ID")
		}
		id, err := ouid.FromObjectIDHex(pictureID)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid picture ID")
		}
		serv := common.MustGetState[*service.Service](ctx, common.StateKeyService)
		picture, err := serv.GetPictureByID(requestCtx, id)
		if errors.Is(err, errs.ErrRecordNotFound) {
			return common.NewError(fiber.StatusNotFound, "picture not found")
		}
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

			pr, pw := io.Pipe()
			streamCtx, cancel := context.WithCancel(ctx.Context())

			errChan := make(chan error, 1)
			streamDetail := *detail
			go func() {
				defer close(errChan)
				err := serv.StorageStreamFile(streamCtx, streamDetail, pw)
				if err != nil {
					pw.CloseWithError(err)
					errChan <- err
					return
				}
				pw.Close()
			}()
			if detail.Mime != "" {
				ctx.Set(fiber.HeaderContentType, detail.Mime)
				return ctx.SendStream(&pictureStream{Reader: pr, pipe: pr, cancel: cancel})
			}

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
					return err
				}
			}

			buf = buf[:n]
			mtype := mimetype.Detect(buf)
			if mtype != nil {
				ctx.Set(fiber.HeaderContentType, mtype.String())
				if updateMIME {
					detectedDetail := streamDetail
					detectedDetail.Mime = mtype.String()
					updatedPicture := *picture
					storageInfo := picture.StorageInfo.Data()
					switch size {
					case "thumb":
						storageInfo.Thumb = &detectedDetail
					case "regular":
						storageInfo.Regular = &detectedDetail
					case "original":
						storageInfo.Original = &detectedDetail
					}
					updatedPicture.StorageInfo = datatypes.NewJSONType(storageInfo)
					go func() {
						if err := serv.SavePicture(context.Background(), &updatedPicture); err != nil {
							log.Errorf("failed to save picture mime: %v", err)
						}
					}()
				}
			}
			ctx.Set(fiber.HeaderContentDisposition, "inline")

			var fullReader io.Reader = bytes.NewReader(buf)
			if err == nil {
				fullReader = io.MultiReader(fullReader, pr)
			}
			return ctx.SendStream(&pictureStream{Reader: fullReader, pipe: pr, cancel: cancel})
		}
		return ctx.Redirect().To(picture.Thumbnail)
	}
}

func HandleGetRandomPicture(ctx fiber.Ctx) error {
	serv := common.MustGetState[*service.Service](ctx, common.StateKeyService)
	picture, err := serv.RandomPictures(ctx.RequestCtx(), 1)
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
