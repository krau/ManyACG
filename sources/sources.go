package sources

import (
	"reflect"
	"strings"

	"github.com/krau/ManyACG/common"
	"github.com/krau/ManyACG/config"
	"github.com/krau/ManyACG/errs"
	sourceCommon "github.com/krau/ManyACG/sources/common"

	"github.com/krau/ManyACG/types"

	_ "github.com/krau/ManyACG/sources/bilibili"
	_ "github.com/krau/ManyACG/sources/danbooru"
	_ "github.com/krau/ManyACG/sources/kemono"
	_ "github.com/krau/ManyACG/sources/nhentai"
	_ "github.com/krau/ManyACG/sources/pixiv"
	_ "github.com/krau/ManyACG/sources/twitter"
	_ "github.com/krau/ManyACG/sources/yandere"
)

func isSourceEnabled(sourceType types.SourceType) bool {
	cfgValue := reflect.ValueOf(config.Cfg.Source)
	field := cfgValue.FieldByName(strings.Title(string(sourceType))) // ignore deprecated warning because we don't need to support unicode punctuation
	if !field.IsValid() {
		return false
	}
	if field.Kind() != reflect.Struct {
		return false
	}
	enableField := field.FieldByName("Enable")
	if !enableField.IsValid() {
		return false
	}
	if enableField.Kind() != reflect.Bool {
		return false
	}
	return enableField.Bool()
}

func InitSources(service types.Service) {
	common.Logger.Info("Initializing sources")
	for sourceType, source := range sourceCommon.Sources {
		if isSourceEnabled(sourceType) {
			source.Init(service)
		} else {
			delete(sourceCommon.Sources, sourceType)
			delete(sourceCommon.SourceURLRegexps, sourceType)
		}
	}
}

func GetSources() map[types.SourceType]types.Source {
	return sourceCommon.Sources
}

func GetArtworkInfo(sourceURL string) (*types.Artwork, error) {
	common.Logger.Infof("Getting artwork info: %s", sourceURL)
	for k, v := range sourceCommon.SourceURLRegexps {
		if v.MatchString(sourceURL) {
			if sourceCommon.Sources[k] != nil {
				return sourceCommon.Sources[k].GetArtworkInfo(sourceURL)
			}
		}
	}
	common.Logger.Warnf("Source URL not supported: %s", sourceURL)
	return nil, errs.ErrSourceNotSupported
}

func GetFileName(artwork *types.Artwork, picture *types.Picture) (string, error) {
	return sourceCommon.GetFileName(artwork, picture)
}

func FindSourceURL(text string) string {
	text = strings.ReplaceAll(text, "\n", " ")
	for sourceType, reg := range sourceCommon.SourceURLRegexps {
		if url := reg.FindString(text); url != "" {
			return sourceCommon.Sources[sourceType].GetCommonSourceURL(url)
		}
	}
	return ""
}

// MatchesSourceURL returns whether the text contains a source URL.
func MatchesSourceURL(text string) bool {
	text = strings.ReplaceAll(text, "\n", " ")
	for _, reg := range sourceCommon.SourceURLRegexps {
		if reg.MatchString(text) {
			return true
		}
	}
	return false
}
