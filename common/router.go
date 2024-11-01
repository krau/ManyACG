package common

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type RestfulCommonResponse[T any] struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	Data    T      `json:"data"`
}

func GinErrorResponse(ctx *gin.Context, err error, status int, message string) {
	if err == nil {
		err = errors.New(message)
	}
	if ctx.GetBool("auth") {
		ctx.JSON(status, &RestfulCommonResponse[any]{
			Status:  status,
			Message: err.Error(),
		})
	} else {
		ctx.JSON(status, &RestfulCommonResponse[any]{
			Status:  status,
			Message: message,
		})
	}
	ctx.Abort()
}

func GinBindError(ctx *gin.Context, err error) {
	GinErrorResponse(ctx, err, http.StatusBadRequest, "Invalid request")
}
