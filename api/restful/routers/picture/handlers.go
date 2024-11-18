package picture

import (
	"bytes"
	"errors"
	"io"
	"net/http"

	"github.com/krau/ManyACG/adapter"
	"github.com/krau/ManyACG/common"
	"github.com/krau/ManyACG/sources"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/krau/ManyACG/service"
	"github.com/krau/ManyACG/storage"
	"github.com/krau/ManyACG/types"

	"github.com/gabriel-vasile/mimetype"
	"github.com/gin-gonic/gin"
)

func GetFile(ctx *gin.Context) {
	picture := ctx.MustGet("picture").(*types.Picture)

	file, err := storage.GetFileStream(ctx, picture.StorageInfo.Original)
	if err != nil {
		common.Logger.Errorf("Failed to get file: %v", err)
		common.GinErrorResponse(ctx, err, http.StatusInternalServerError, "Failed to get file")
		return
	}
	defer file.Close()

	var mimeBuf bytes.Buffer
	tee := io.TeeReader(io.LimitReader(file, 512), &mimeBuf)

	_, err = io.Copy(io.Discard, tee)
	if err != nil {
		common.Logger.Errorf("Failed to read for mime detection: %v", err)
		common.GinErrorResponse(ctx, err, http.StatusInternalServerError, "Failed to process file")
		return
	}

	mimeType := mimetype.Detect(mimeBuf.Bytes())
	ctx.Header("Content-Type", mimeType.String())
	artworkID, _ := primitive.ObjectIDFromHex(picture.ArtworkID)
	artwork, err := service.GetArtworkByID(ctx, artworkID, adapter.LoadNone())
	if err != nil {
		common.Logger.Errorf("Failed to get artwork: %v", err)
		common.GinErrorResponse(ctx, err, http.StatusInternalServerError, "Failed to get artwork")
		return
	}
	fileName, _ := sources.GetFileName(artwork, picture)
	if fileName != "" {
		ctx.Header("Content-Disposition", "inline; filename="+fileName)
	}

	if _, err := io.Copy(ctx.Writer, &mimeBuf); err != nil {
		common.Logger.Errorf("Failed to write mime buffer: %v", err)
		common.GinErrorResponse(ctx, err, http.StatusInternalServerError, "Failed to send file")
		return
	}

	_, err = io.Copy(ctx.Writer, file)
	if err != nil {
		common.Logger.Errorf("Failed to copy file: %v", err)
		common.GinErrorResponse(ctx, err, http.StatusInternalServerError, "Failed to send file")
		return
	}
}

func RandomPicture(ctx *gin.Context) {
	pictures, err := service.GetRandomPictures(ctx, 1)

	if err != nil {
		common.Logger.Errorf("Failed to get random pictures: %v", err)
		common.GinErrorResponse(ctx, err, http.StatusInternalServerError, "Failed to get random pictures")
		return
	}
	if len(pictures) == 0 {
		common.GinErrorResponse(ctx, errors.New("not found pictures"), http.StatusNotFound, "Pictures not found")
		return
	}
	picture := pictures[0]
	switch picture.StorageInfo.Regular.Type {
	case types.StorageTypeLocal:
		ctx.File(picture.StorageInfo.Regular.Path)
	case types.StorageTypeAlist:
		ctx.Redirect(http.StatusFound, common.ApplyPathRule(picture.StorageInfo.Regular.Path))
	default:
		data, err := storage.GetFile(ctx, picture.StorageInfo.Regular)
		if err != nil {
			common.Logger.Errorf("Failed to get file: %v", err)
			common.GinErrorResponse(ctx, err, http.StatusInternalServerError, "Failed to get file")
			return
		}
		mimeType := mimetype.Detect(data)
		ctx.Data(http.StatusOK, mimeType.String(), data)
	}
}
