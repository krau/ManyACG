package kemono

import (
	"errors"
	"fmt"
	"regexp"
)

var (
	kemonoSourceURLRegex = regexp.MustCompile(`kemono\.(su|cr)/\w+/user/\d+/post/\d+`)
	kemonoDomainBase     = "https://kemono.cr"
	apiBaseURL           = fmt.Sprintf("%s/api/v1", kemonoDomainBase)
	thumbnailsBase       = "https://img.kemono.cr/thumbnail/data"
)

var (
	ErrIndexOOB             = errors.New("index out of artwork pictures bounds")
	ErrInvalidKemonoPostURL = errors.New("invalid kemono post url")
	ErrNotPicture           = errors.New("kemono post files or attachments are not pictures")
)
