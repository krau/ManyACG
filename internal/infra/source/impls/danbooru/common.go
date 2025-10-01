package danbooru

import (
	"errors"
	"regexp"

	"github.com/krau/ManyACG/types"
)

var (
	danbooruSourceURLRegexp = regexp.MustCompile(`danbooru\.donmai\.us/(posts|post\/show)/\d+`)
	fakeArtist              = &types.Artist{
		Name:     "Danbooru",
		Username: "Danbooru",
		UID:      "1",
		Type:     types.SourceTypeDanbooru,
	}
	ErrInvalidDanbooruPostURL = errors.New("invalid danbooru post url")
	numberRegexp              = regexp.MustCompile(`\d+`)
)

func GetPostID(url string) string {
	matchUrl := danbooruSourceURLRegexp.FindString(url)
	if matchUrl == "" {
		return ""
	}
	return numberRegexp.FindString(matchUrl)
}
