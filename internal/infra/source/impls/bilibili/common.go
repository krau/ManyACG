package bilibili

import (
	"errors"
	"regexp"
)

var (
	dynamicURLRegexp           *regexp.Regexp = regexp.MustCompile(`t.bilibili.com/(\d+)|bilibili.com/opus/(\d+)`)
	webDynamicAPIURLFormat     string         = "https://api.bilibili.com/x/polymer/web-dynamic/v1/detail?id=%s"
	desktopDynamicAPIURLFormat string         = "https://api.bilibili.com/x/polymer/web-dynamic/desktop/v1/detail?id=%s"
)

var (
	ErrRequestFailed = errors.New("request bilibili dynamic url failed")
	ErrIndexOOB      = errors.New("index out of artwork pictures bounds")
	ErrInvalidURL    = errors.New("invalid bilibili dynamic url")
)
