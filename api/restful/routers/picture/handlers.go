package picture

import (
	"ManyACG/common"
	. "ManyACG/logger"
	"ManyACG/storage"
	"ManyACG/types"
	"net/http"

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
