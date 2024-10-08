package twitter

import (
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/krau/ManyACG/config"
	"github.com/krau/ManyACG/types"
)

type Twitter struct{}

func (t *Twitter) Init() {
}

func (t *Twitter) FetchNewArtworksWithCh(artworkCh chan *types.Artwork, limit int) error {
	return nil
}

func (t *Twitter) FetchNewArtworks(limit int) ([]*types.Artwork, error) {
	return nil, nil
}

func (t *Twitter) GetArtworkInfo(sourceURL string) (*types.Artwork, error) {
	tweetPath := GetTweetPath(sourceURL)
	if tweetPath == "" {
		return nil, ErrInvalidURL
	}
	fxTwitterApiURL := "https://api." + config.Cfg.Source.Twitter.FxTwitterDomain + "/" + tweetPath
	resp, err := reqApiResp(fxTwitterApiURL)
	if err != nil {
		return nil, err
	}
	return resp.ToArtwork()
}

func (t *Twitter) GetPictureInfo(sourceURL string, index uint) (*types.Picture, error) {
	artwork, err := t.GetArtworkInfo(sourceURL)
	if err != nil {
		return nil, err
	}
	if index >= uint(len(artwork.Pictures)) {
		return nil, ErrIndexOOB
	}
	return artwork.Pictures[index], nil
}

func (t *Twitter) GetSourceURLRegexp() *regexp.Regexp {
	return twitterSourceURLRegexp
}

func (t *Twitter) GetCommonSourceURL(url string) string {
	tweetPath := GetTweetPath(url)
	if tweetPath == "" {
		return ""
	}
	return "https://twitter.com/" + tweetPath
}

func (t *Twitter) GetFileName(artwork *types.Artwork, picture *types.Picture) string {
	original := picture.Original
	urlSplit := strings.Split(picture.Original, "?")
	if len(urlSplit) > 1 {
		original = strings.Join(urlSplit[:len(urlSplit)-1], "?")
	}
	tweetID := strings.Split(artwork.SourceURL, "/")[len(strings.Split(artwork.SourceURL, "/"))-1]
	return tweetID + "_" + strconv.Itoa(int(picture.Index)) + filepath.Ext(original)
}

func (t *Twitter) Config() *config.SourceCommonConfig {
	return &config.SourceCommonConfig{
		Enable:   config.Cfg.Source.Twitter.Enable,
		Intervel: -1, // Twitter 暂无法实现主动抓取的功能
	}

}
