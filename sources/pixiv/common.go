package pixiv

import (
	"errors"
	"regexp"

	"github.com/imroc/req/v3"
)

var reqClient *req.Client

var (
	pixivSourceURLRegexp  *regexp.Regexp = regexp.MustCompile(`pixiv\.net/(?:artworks/|i/|member_illust\.php\?(?:[\w=&]*\&|)illust_id=)(\d+)`)
	numberRegexp          *regexp.Regexp = regexp.MustCompile(`\d+`)
	ErrUnmarshalPixivAjax                = errors.New("error decoding artwork info, maybe the artwork is deleted")
)
