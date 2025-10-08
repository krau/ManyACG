package picture

import (
	"errors"
	"net/http"

	"github.com/krau/ManyACG/api/restful/utils"
	"github.com/krau/ManyACG/common"

	"github.com/krau/ManyACG/internal/service"
	"github.com/krau/ManyACG/storage"
	"github.com/krau/ManyACG/types"

	"github.com/gabriel-vasile/mimetype"
	"github.com/gin-gonic/gin"
)

func GetFile(ctx *gin.Context) {
	picture := ctx.MustGet("picture").(*types.Picture)
	var data []byte
	var err error
	if picture.StorageInfo.Original != nil {
		data, err = storage.GetFile(ctx, picture.StorageInfo.Original)
	} else {
		data, err = common.DownloadWithCache(ctx, picture.Original, nil)
	}
	if err != nil {
		common.Logger.Errorf("Failed to get file: %v", err)
		utils.GinErrorResponse(ctx, err, http.StatusInternalServerError, "Failed to get file")
		return
	}
	mimeType := mimetype.Detect(data)
	ctx.Data(http.StatusOK, mimeType.String(), data)

	// GetFileStream:
	//
	// file, err := storage.GetFileStream(ctx, picture.StorageInfo.Original)
	// if err != nil {
	// 	common.Logger.Errorf("Failed to get file: %v", err)
	// 	utils.GinErrorResponse(ctx, err, http.StatusInternalServerError, "Failed to get file")
	// 	return
	// }
	// defer file.Close()

	// var mimeBuf bytes.Buffer
	// tee := io.TeeReader(io.LimitReader(file, 512), &mimeBuf)

	// _, err = io.Copy(io.Discard, tee)
	// if err != nil {
	// 	common.Logger.Errorf("Failed to read for mime detection: %v", err)
	// 	utils.GinErrorResponse(ctx, err, http.StatusInternalServerError, "Failed to process file")
	// 	return
	// }

	// mimeType := mimetype.Detect(mimeBuf.Bytes())
	// ctx.Header("Content-Type", mimeType.String())
	// artworkID, _ := primitive.ObjectIDFromHex(picture.ArtworkID)
	// artwork, err := service.GetArtworkByID(ctx, artworkID, adapter.LoadNone())
	// if err != nil {
	// 	common.Logger.Errorf("Failed to get artwork: %v", err)
	// 	utils.GinErrorResponse(ctx, err, http.StatusInternalServerError, "Failed to get artwork")
	// 	return
	// }
	// fileName, _ := sources.GetFileName(artwork, picture)
	// if fileName != "" {
	// 	ctx.Header("Content-Disposition", "inline; filename="+fileName)
	// }

	// if _, err := io.Copy(ctx.Writer, &mimeBuf); err != nil {
	// 	common.Logger.Errorf("Failed to write mime buffer: %v", err)
	// 	utils.GinErrorResponse(ctx, err, http.StatusInternalServerError, "Failed to send file")
	// 	return
	// }

	// _, err = io.Copy(ctx.Writer, file)
	// if err != nil {
	// 	common.Logger.Errorf("Failed to copy file: %v", err)
	// 	utils.GinErrorResponse(ctx, err, http.StatusInternalServerError, "Failed to send file")
	// 	return
	// }
}

func RandomPicture(ctx *gin.Context) {
	pictures, err := service.GetRandomPictures(ctx, 1)

	if err != nil {
		common.Logger.Errorf("Failed to get random pictures: %v", err)
		utils.GinErrorResponse(ctx, err, http.StatusInternalServerError, "Failed to get random pictures")
		return
	}
	if len(pictures) == 0 {
		utils.GinErrorResponse(ctx, errors.New("not found pictures"), http.StatusNotFound, "Pictures not found")
		return
	}
	picture := pictures[0]
	if picture.StorageInfo == nil || picture.StorageInfo.Regular == nil {
		ctx.Redirect(http.StatusFound, picture.Thumbnail)
		return
	}
	picUrl := utils.ApplyApiStoragePathRule(picture.StorageInfo.Regular)
	if picUrl == "" || picUrl == picture.StorageInfo.Regular.Path {
		storage.ServeFile(ctx, picture.StorageInfo.Regular)
		return
	}
	ctx.Redirect(http.StatusFound, picUrl)
}
