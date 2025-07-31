package kemono

import (
	"errors"
	"regexp"

	"github.com/imroc/req/v3"
)

var (
	reqClient            *req.Client
	kemonoSourceURLRegex = regexp.MustCompile(`kemono\.(su|cr)/\w+/user/\d+/post/\d+`)
	apiBaseURL           = "https://kemono.cr/api/v1"
	cdnBaseURL           = "https://c1.kemono.cr/data"
	thumbnailsBaseURL    = "https://img.kemono.cr/thumbnail/data"
)

var (
	ErrIndexOOB             = errors.New("index out of artwork pictures bounds")
	ErrInvalidKemonoPostURL = errors.New("invalid kemono post url")
	ErrNotPicture           = errors.New("kemono post files or attachments are not pictures")
)
