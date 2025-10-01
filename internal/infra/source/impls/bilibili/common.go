package bilibili

import (
	"errors"
	"regexp"

	"github.com/imroc/req/v3"
)

var reqClient *req.Client

var (
	dynamicURLRegexp           *regexp.Regexp = regexp.MustCompile(`t.bilibili.com/(\d+)|bilibili.com/opus/(\d+)`)
	numberRegexp               *regexp.Regexp = regexp.MustCompile(`\d+`)
	webDynamicAPIURLFormat     string         = "https://api.bilibili.com/x/polymer/web-dynamic/v1/detail?id=%s"
	desktopDynamicAPIURLFormat string         = "https://api.bilibili.com/x/polymer/web-dynamic/desktop/v1/detail?id=%s"
)

var (
	ErrRequestFailed = errors.New("request bilibili dynamic url failed")
	ErrIndexOOB      = errors.New("index out of artwork pictures bounds")
	ErrInvalidURL    = errors.New("invalid bilibili dynamic url")
)
