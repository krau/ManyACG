package utils

import (
	"errors"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/krau/ManyACG/common"
	"github.com/krau/ManyACG/config"
	"github.com/krau/ManyACG/types"
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

// 在api返回中重写图片路径, 用于拼接图片直链
//
// 例: rule.Type = "alist", rule.Path = "/pictures/", rule.JoinPrefix = "https://example.com/pictures/", rule.TrimPrefix = "/pictures/"
//
// -> https://example.com/pictures/1234567890abcdef.jpg
//
// 如果图片路径以rule.Path开头, 且rule.StorageType为空或与图片存储类型匹配, 则将图片路径转换为rule.JoinPrefix + 图片路径去掉rule.TrimPrefix的部分
//
// 否则返回原图片路径
func ApplyApiStoragePathRule(detail *types.StorageDetail) string {
	for _, rule := range config.Cfg.API.PathRules {
		if strings.HasPrefix(detail.Path, rule.Path) && (rule.StorageType == "" || rule.StorageType == string(detail.Type)) {
			parsedUrl, err := url.JoinPath(rule.JoinPrefix, strings.TrimPrefix(detail.Path, rule.TrimPrefix))
			if err != nil {
				common.Logger.Warnf("Failed to join path: %s", err)
				return detail.Path
			}
			return parsedUrl
		}
	}
	return detail.Path
}
