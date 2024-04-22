package sources

import (
	"ManyACG-Bot/common"
	"ManyACG-Bot/config"
	"ManyACG-Bot/errors"
	"ManyACG-Bot/sources/pixiv"
	"ManyACG-Bot/types"
)

var Sources = make(map[string]Source)

func init() {
	if config.Cfg.Source.Pixiv.Enable {
		Sources["pixiv"] = new(pixiv.Pixiv)
	}
}

func GetArtworkInfo(sourceURL string) (*types.Artwork, error) {
	for k, v := range common.SourceURLRegexps {
		if v.MatchString(sourceURL) {
			if Sources[k] != nil {
				return Sources[k].GetArtworkInfo(sourceURL)
			}
		}
	}
	return nil, errors.ErrSourceNotSupported
}
