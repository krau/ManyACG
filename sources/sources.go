package sources

import (
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
	for _, source := range Sources {
		artwork, err := source.GetArtworkInfo(sourceURL)
		if err == nil {
			return artwork, nil
		}
	}
	return nil, errors.ErrSourceNotSupported
}
