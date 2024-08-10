package kemono

import (
	"errors"
	"regexp"

	"github.com/imroc/req/v3"
)

var (
	kemonoSourceURLRegex = regexp.MustCompile(`kemono\.su/\w+/user/\d+/post/\d+`)
	reqClient            = req.NewClient()
	apiBaseURL           = "https://kemono.su/api/v1"
	cdnBaseURL           = "https://c1.kemono.su/data"
	thumbnailsBaseURL    = "https://img.kemono.su/thumbnail/data"
)

var (
	ErrIndexOOB             = errors.New("index out of artwork pictures bounds")
	ErrInvalidKemonoPostURL = errors.New("invalid kemono post url")
	ErrNotPicture           = errors.New("kemono post files or attachments are not pictures")
)

/*
https://n2.kemono.su/data/e6/fe/e6fe64943e79d530ea01659e3601b70ed81918a286ae032e6408054d6ac3fd0f.jpg
https://img.kemono.su/thumbnail/data/e6/fe/e6fe64943e79d530ea01659e3601b70ed81918a286ae032e6408054d6ac3fd0f.jpg
*/
