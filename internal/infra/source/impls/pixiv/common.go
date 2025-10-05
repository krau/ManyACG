package pixiv

import (
	"errors"
	"regexp"
)

var (
	sourceReg             = regexp.MustCompile(`pixiv\.net/(?:en/)?(?:artworks/|i/|member_illust\.php\?(?:[\w=&]*\&|)illust_id=)(\d+)`)
	ErrUnmarshalPixivAjax = errors.New("error decoding artwork info, maybe the artwork is deleted")
)
