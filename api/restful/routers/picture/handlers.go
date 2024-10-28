package picture

import (
	"errors"
	"net/http"

	"github.com/krau/ManyACG/common"
	. "github.com/krau/ManyACG/logger"
	"github.com/krau/ManyACG/service"
	"github.com/krau/ManyACG/storage"
	"github.com/krau/ManyACG/types"

	"github.com/gabriel-vasile/mimetype"
	"github.com/gin-gonic/gin"
)

func GetFile(ctx *gin.Context) {
	picture := ctx.MustGet("picture").(*types.Picture)
	data, err := storage.GetFile(ctx, picture.StorageInfo.Original)
	if err != nil {
		Logger.Errorf("Failed to get file: %v", err)
		common.GinErrorResponse(ctx, err, http.StatusInternalServerError, "Failed to get file")
		return
	}
	mimeType := mimetype.Detect(data)
	ctx.Data(http.StatusOK, mimeType.String(), data)
}

func RandomPicture(ctx *gin.Context) {
	pictures, err := service.GetRandomPictures(ctx, 1)

	if err != nil {
		Logger.Errorf("Failed to get random pictures: %v", err)
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
			Logger.Errorf("Failed to get file: %v", err)
			common.GinErrorResponse(ctx, err, http.StatusInternalServerError, "Failed to get file")
			return
		}
		mimeType := mimetype.Detect(data)
		ctx.Data(http.StatusOK, mimeType.String(), data)
	}
}
