package picture

import (
	. "ManyACG/logger"
	"ManyACG/storage"
	"ManyACG/types"
	"net/http"

	"github.com/gabriel-vasile/mimetype"
	"github.com/gin-gonic/gin"
)

func GetThumb(ctx *gin.Context) {
	picture := ctx.MustGet("picture").(*types.Picture)
	ctx.Redirect(http.StatusFound, picture.Thumbnail)
}

func GetFile(ctx *gin.Context) {
	picture := ctx.MustGet("picture").(*types.Picture)
	data, err := storage.GetFile(ctx, picture.StorageInfo)
	if err != nil {
		Logger.Errorf("Failed to get file: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status": http.StatusInternalServerError,
			"message": func() string {
				if ctx.GetBool("auth") {
					return err.Error()
				}
				return "Failed to get file"
			}(),
		})
		return
	}
	mimeType := mimetype.Detect(data)
	ctx.Data(http.StatusOK, mimeType.String(), data)
}
