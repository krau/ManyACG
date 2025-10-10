package danbooru

import (
	"errors"
	"regexp"

	"github.com/krau/ManyACG/internal/model/dto"
	"github.com/krau/ManyACG/internal/shared"
	"github.com/krau/ManyACG/pkg/reutil"
)

var (
	danbooruSourceURLRegexp = regexp.MustCompile(`danbooru\.donmai\.us/(posts|post\/show)/\d+`)
	fakeArtist              = &dto.FetchedArtist{
		Name:     "Danbooru",
		Username: "Danbooru",
		UID:      "1",
		Type:     shared.SourceTypeDanbooru,
	}
	ErrInvalidDanbooruPostURL = errors.New("invalid danbooru post url")
)

func GetPostID(url string) string {
	matchUrl := danbooruSourceURLRegexp.FindString(url)
	if matchUrl == "" {
		return ""
	}
	id, ok := reutil.GetLatestNumberFromString(matchUrl)
	if !ok {
		return ""
	}
	return id
}
