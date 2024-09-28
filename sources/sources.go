package sources

import (
	"regexp"
	"strings"

	"github.com/krau/ManyACG/common"
	"github.com/krau/ManyACG/config"
	"github.com/krau/ManyACG/errors"
	. "github.com/krau/ManyACG/logger"
	"github.com/krau/ManyACG/sources/bilibili"
	"github.com/krau/ManyACG/sources/danbooru"
	"github.com/krau/ManyACG/sources/kemono"
	"github.com/krau/ManyACG/sources/pixiv"
	"github.com/krau/ManyACG/sources/twitter"
	"github.com/krau/ManyACG/types"
)

var (
	Sources          = make(map[types.SourceType]Source)
	SourceURLRegexps = make(map[types.SourceType]*regexp.Regexp)
)

var (
	sourcesNotInit = make(map[types.SourceType]Source)
)

func newSource(sourceType types.SourceType) Source {
	switch sourceType {
	case types.SourceTypePixiv:
		return new(pixiv.Pixiv)
	case types.SourceTypeTwitter:
		return new(twitter.Twitter)
	case types.SourceTypeBilibili:
		return new(bilibili.Bilibili)
	case types.SourceTypeKemono:
		return new(kemono.Kemono)
	case types.SourceTypeDanbooru:
		return new(danbooru.Danbooru)
	}
	return nil
}

func InitSources() {
	Logger.Info("Initializing sources")
	for _, sourceType := range types.SourceTypes {
		sourcesNotInit[sourceType] = newSource(sourceType)
	}
	if config.Cfg.Source.Pixiv.Enable {
		Sources[types.SourceTypePixiv] = sourcesNotInit[types.SourceTypePixiv]
		Sources[types.SourceTypePixiv].Init()
	}
	if config.Cfg.Source.Twitter.Enable {
		Sources[types.SourceTypeTwitter] = sourcesNotInit[types.SourceTypeTwitter]
		Sources[types.SourceTypeTwitter].Init()
	}
	if config.Cfg.Source.Bilibili.Enable {
		Sources[types.SourceTypeBilibili] = sourcesNotInit[types.SourceTypeBilibili]
		Sources[types.SourceTypeBilibili].Init()
	}
	if config.Cfg.Source.Kemono.Enable {
		Sources[types.SourceTypeKemono] = sourcesNotInit[types.SourceTypeKemono]
		Sources[types.SourceTypeKemono].Init()
	}
	if config.Cfg.Source.Danbooru.Enable {
		Sources[types.SourceTypeDanbooru] = sourcesNotInit[types.SourceTypeDanbooru]
		Sources[types.SourceTypeDanbooru].Init()
	}
	for sourceType, source := range Sources {
		SourceURLRegexps[sourceType] = source.GetSourceURLRegexp()
	}
}

func GetArtworkInfo(sourceURL string) (*types.Artwork, error) {
	Logger.Infof("Getting artwork info: %s", sourceURL)
	for k, v := range SourceURLRegexps {
		if v.MatchString(sourceURL) {
			if Sources[k] != nil {
				return Sources[k].GetArtworkInfo(sourceURL)
			}
		}
	}
	Logger.Warnf("Source URL not supported: %s", sourceURL)
	return nil, errors.ErrSourceNotSupported
}

func GetFileName(artwork *types.Artwork, picture *types.Picture) (string, error) {
	if sourcesNotInit[artwork.SourceType] == nil {
		return "", errors.ErrSourceNotSupported
	}
	fileName := sourcesNotInit[artwork.SourceType].GetFileName(artwork, picture)
	return common.EscapeFileName(fileName), nil
}

func FindSourceURL(text string) string {
	text = strings.ReplaceAll(text, "\n", " ")
	for sourceType, reg := range SourceURLRegexps {
		if url := reg.FindString(text); url != "" {
			return Sources[sourceType].GetCommonSourceURL(url)
		}
	}
	return ""
}

// MatchesSourceURL returns whether the text contains a source URL.
func MatchesSourceURL(text string) bool {
	text = strings.ReplaceAll(text, "\n", " ")
	for _, reg := range SourceURLRegexps {
		if reg.MatchString(text) {
			return true
		}
	}
	return false
}
