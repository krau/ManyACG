package utils

import "github.com/gin-gonic/gin"

func BindError(ctx *gin.Context, err error) string {
	if ctx.GetBool("auth") {
		return err.Error()
	}
	return "Invalid request"
}
