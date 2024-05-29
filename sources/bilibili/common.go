package bilibili

import (
	"errors"
	"regexp"

	"github.com/imroc/req/v3"
)

var ReqClient *req.Client

var (
	dynamicURLRegexp *regexp.Regexp = regexp.MustCompile(`t.bilibili.com/(\d+)|m.bilibili.com/opus/(\d+)`)
	numberRegexp     *regexp.Regexp = regexp.MustCompile(`\d+`)
	apiURLFormat     string         = "https://api.bilibili.com/x/polymer/web-dynamic/v1/detail?timezone_offset=-480&platform=web&id=%s&features=itemOpusStyle"
)

var (
	ErrRequestFailed = errors.New("request bilibili dynamic url failed")
	ErrIndexOOB      = errors.New("index out of artwork pictures bounds")
	ErrInvalidURL    = errors.New("invalid bilibili dynamic url")
)
