package picture

import (
	. "ManyACG/logger"
	"ManyACG/storage"
	"ManyACG/types"
	"context"
	"net/http"
	"time"

	"github.com/gabriel-vasile/mimetype"
	"github.com/gin-gonic/gin"
)

func GetThumb(ctx *gin.Context) {
	picture := ctx.MustGet("picture").(*types.Picture)
	ctx.Redirect(http.StatusFound, picture.Thumbnail)
}
func GetFile(ctx *gin.Context) {
	picture := ctx.MustGet("picture").(*types.Picture)
	timeoutCtx, cancel := context.WithTimeout(ctx.Request.Context(), 1*time.Second)
	defer cancel()

	resultChan := make(chan struct {
		data []byte
		err  error
	})

	go func() {
		data, err := storage.GetFile(picture.StorageInfo)
		resultChan <- struct {
			data []byte
			err  error
		}{data, err}
	}()

	select {
	case result := <-resultChan:
		if result.err != nil {
			Logger.Errorf("Failed to get file: %v", result.err)
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"status":  http.StatusInternalServerError,
				"message": "Failed to get file",
			})
			return
		}
		mimeType := mimetype.Detect(result.data)
		ctx.Data(http.StatusOK, mimeType.String(), result.data)
	case <-timeoutCtx.Done():
		Logger.Warnf("File download timed out")
		ctx.JSON(http.StatusRequestTimeout, gin.H{
			"status":  http.StatusRequestTimeout,
			"message": "File download timed out",
		})
	}
}
