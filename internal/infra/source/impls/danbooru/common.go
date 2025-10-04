package danbooru

import (
	"errors"
	"regexp"

	"github.com/krau/ManyACG/internal/model/dto"
	"github.com/krau/ManyACG/internal/shared"
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
	numberRegexp              = regexp.MustCompile(`\d+`)
)

func GetPostID(url string) string {
	matchUrl := danbooruSourceURLRegexp.FindString(url)
	if matchUrl == "" {
		return ""
	}
	return numberRegexp.FindString(matchUrl)
}
