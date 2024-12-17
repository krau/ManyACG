package yandere

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"regexp"
	"strconv"

	"github.com/imroc/req/v3"
	"github.com/krau/ManyACG/common"
	"github.com/krau/ManyACG/config"
	sourceCommon "github.com/krau/ManyACG/sources/common"
	"github.com/krau/ManyACG/types"
)

type Yandere struct{}

var reqClient *req.Client

var service types.Service

func init() {
	sourceCommon.RegisterSource(types.SourceTypeYandere, new(Yandere))
}

func (y *Yandere) Init(s types.Service) {
	reqClient = req.C().ImpersonateChrome().SetCommonRetryCount(2)
	if config.Cfg.Source.Proxy != "" {
		reqClient.SetProxyURL(config.Cfg.Source.Proxy)
	}
	service = s
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
	common.Logger.Tracef("getting yandere api: %s", apiBaseURL+postID)
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
	parentID := 0
	if len(yandereResp) == 1 {
		parentID = yandereResp[0].ParentID
	}
	if parentID == 0 {
		return yandereResp.ToArtwork(), nil
	}

	parentURL := sourceURLPrefix + strconv.Itoa(parentID)
	artwork, _ := service.GetArtworkByURL(context.TODO(), parentURL)
	if artwork != nil {
		return artwork, nil
	}

	apiURL := apiBaseURL + strconv.Itoa(parentID)
	common.Logger.Tracef("getting yandere api: %s", apiURL)
	resp, err = reqClient.R().Get(apiURL)
	if err != nil {
		return nil, err
	}
	if resp.IsErrorState() {
		return nil, fmt.Errorf("%w: %s", ErrYandereAPIError, resp.Status)
	}
	var parentResp YandereJsonResp
	if err := json.Unmarshal(resp.Bytes(), &parentResp); err != nil {
		return nil, err
	}

	return parentResp.ToArtwork(), nil
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
		idStr = picture.ID
	}
	if idStr == "" {
		idStr = common.MD5Hash(picture.Original)
	}
	return "yandere_" + idStr + "_" + strconv.Itoa(int(picture.Index)) + filepath.Ext(picture.Original)
}

func (y *Yandere) Config() *config.SourceCommonConfig {
	return &config.SourceCommonConfig{
		Enable:   config.Cfg.Source.Yandere.Enable,
		Intervel: -1,
	}
}
