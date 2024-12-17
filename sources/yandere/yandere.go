package yandere

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"regexp"

	"github.com/imroc/req/v3"
	"github.com/krau/ManyACG/common"
	"github.com/krau/ManyACG/config"
	sourceCommon "github.com/krau/ManyACG/sources/common"
	"github.com/krau/ManyACG/types"
)

type Yandere struct{}

var reqClient *req.Client

func init() {
	sourceCommon.RegisterSource(types.SourceTypeYandere, new(Yandere))
}

func (y *Yandere) Init(service types.Service) {
	reqClient = req.C().ImpersonateChrome().SetCommonRetryCount(2)
	if config.Cfg.Source.Proxy != "" {
		reqClient.SetProxyURL(config.Cfg.Source.Proxy)
	}

}

func (y *Yandere) FetchNewArtworksWithCh(artworkCh chan *types.Artwork, limit int) error {
	return nil
}

func (y *Yandere) FetchNewArtworks(limit int) ([]*types.Artwork, error) {
	return nil, nil
}

func (y *Yandere) GetArtworkInfo(sourceURL string) (*types.Artwork, error) {
	postID := GetPostID(sourceURL)
	if postID == "" {
		return nil, ErrInvalidYanderePostURL
	}
	sourceURL = sourceURLPrefix + postID
	common.Logger.Tracef("request artwork info: %s", sourceURL)
	resp, err := reqClient.R().Get(apiBaseURL + postID)
	if err != nil {
		return nil, err
	}
	if resp.IsErrorState() {
		return nil, fmt.Errorf("%w: %s", ErrYandereAPIError, resp.Status)
	}
	var yandereResp YandereJsonResp
	if err := json.Unmarshal(resp.Bytes(), &yandereResp); err != nil {
		return nil, err
	}
	return yandereResp.ToArtwork(), nil
}

func (y *Yandere) GetPictureInfo(sourceURL string, _ uint) (*types.Picture, error) {
	a, err := y.GetArtworkInfo(sourceURL)
	if err != nil {
		return nil, err
	}
	return a.Pictures[0], nil
}

func (y *Yandere) GetSourceURLRegexp() *regexp.Regexp {
	return yandereSourceURLRegexp
}

func (y *Yandere) GetCommonSourceURL(url string) string {
	postID := GetPostID(url)
	if postID == "" {
		return ""
	}
	return sourceURLPrefix + postID
}

func (y *Yandere) GetFileName(artwork *types.Artwork, picture *types.Picture) string {
	idStr := GetPostID(artwork.SourceURL)
	if idStr == "" {
		idStr = "unknown"
	}
	return "yandere_" + idStr + filepath.Ext(picture.Original)
}

func (y *Yandere) Config() *config.SourceCommonConfig {
	return &config.SourceCommonConfig{
		Enable:   config.Cfg.Source.Yandere.Enable,
		Intervel: -1,
	}
}
