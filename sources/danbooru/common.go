package danbooru

import (
	"ManyACG/types"
	"errors"
	"regexp"
)

var (
	danbooruSourceURLRegexp = regexp.MustCompile(`danbooru\.donmai\.us/posts/\d+`)
	fakeArtist              = &types.Artist{
		Name:     "Danbooru",
		Username: "Danbooru",
		UID:      "1",
		Type:     types.SourceTypeDanbooru,
	}
	ErrInvalidDanbooruPostURL = errors.New("invalid danbooru post url")
)
