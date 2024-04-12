package sources

import (
	"ManyACG-Bot/config"
	"ManyACG-Bot/sources/pixiv"
)

var Sources = make(map[string]Source)

func init() {
	if config.Cfg.Source.Pixiv.Enable {
		Sources["pixiv"] = new(pixiv.Pixiv)
	}
}
