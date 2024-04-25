package sources

import (
	"ManyACG-Bot/common"
	"ManyACG-Bot/config"
	"ManyACG-Bot/errors"
	"ManyACG-Bot/sources/pixiv"
	"ManyACG-Bot/sources/twitter"
	"ManyACG-Bot/types"
	"path/filepath"
	"strconv"
	"strings"
)

var Sources = make(map[string]Source)

func init() {
	if config.Cfg.Source.Pixiv.Enable {
		Sources["pixiv"] = new(pixiv.Pixiv)
	}
	if config.Cfg.Source.Twitter.Enable {
		Sources["twitter"] = new(twitter.Twitter)
	}
}

func GetArtworkInfo(sourceURL string) (*types.Artwork, error) {
	for k, v := range SourceURLRegexps {
		if v.MatchString(sourceURL) {
			if Sources[k] != nil {
				return Sources[k].GetArtworkInfo(sourceURL)
			}
		}
	}
	return nil, errors.ErrSourceNotSupported
}

func GetFileName(artwork *types.Artwork, picture *types.Picture) string {
	fileName := ""
	switch artwork.SourceType {
	case types.SourceTypePixiv:
		fileName = artwork.Title + "_" + filepath.Base(picture.Original)
	case types.SourceTypeTwitter:
		original := picture.Original
		urlSplit := strings.Split(picture.Original, "?")
		if len(urlSplit) > 1 {
			original = strings.Join(urlSplit[:len(urlSplit)-1], "?")
		}
		tweetID := strings.Split(artwork.SourceURL, "/")[len(strings.Split(artwork.SourceURL, "/"))-1]
		fileName = tweetID + "_" + strconv.Itoa(int(picture.Index)) + filepath.Ext(original)
	default:
		fileName = artwork.Title + "_" + filepath.Base(picture.Original)
	}
	return common.ReplaceFileNameInvalidChar(fileName)
}
