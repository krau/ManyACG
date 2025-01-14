package kemono

import (
	"encoding/json"
	"net/http"
	"path/filepath"
	"regexp"
	"strconv"
	"time"

	"github.com/imroc/req/v3"
	"github.com/krau/ManyACG/common"
	"github.com/krau/ManyACG/config"
	sourceCommon "github.com/krau/ManyACG/sources/common"

	"github.com/krau/ManyACG/types"
)

type Kemono struct{}

func init() {
	sourceCommon.RegisterSource(types.SourceTypeKemono, new(Kemono))
}

func (k *Kemono) Init(_ types.Service) {
	reqClient = req.C().ImpersonateChrome()
	if config.Cfg.Source.Kemono.Session != "" {
		reqClient.SetCommonCookies(&http.Cookie{
			Name:  "session",
			Value: config.Cfg.Source.Kemono.Session,
		})
	}
	if config.Cfg.Source.Proxy != "" {
		reqClient.SetProxyURL(config.Cfg.Source.Proxy)
	}
	reqClient.SetCommonRetryCount(3).SetTLSHandshakeTimeout(20 * time.Second)
}

func (k *Kemono) FetchNewArtworksWithCh(artworkCh chan *types.Artwork, limit int) error {
	return nil
}

func (k *Kemono) FetchNewArtworks(limit int) ([]*types.Artwork, error) {
	return nil, nil
}

func (k *Kemono) GetArtworkInfo(sourceURL string) (*types.Artwork, error) {
	kemonoPostURL := kemonoSourceURLRegex.FindString(sourceURL)
	if kemonoPostURL == "" {
		return nil, ErrInvalidKemonoPostURL
	}
	sourceURL = "https://" + kemonoPostURL
	postPath := getPostPath(sourceURL)
	apiURL := apiBaseURL + postPath
	common.Logger.Tracef("request artwork info: %s", apiURL)
	resp, err := reqClient.R().Get(apiURL)
	if err != nil {
		return nil, err
	}
	var kemonoResp KemonoPostResp
	if err := json.Unmarshal(resp.Bytes(), &kemonoResp); err != nil {
		return nil, err
	}
	return kemonoResp.ToArtwork()
}

func (k *Kemono) GetPictureInfo(sourceURL string, index uint) (*types.Picture, error) {
	artwork, err := k.GetArtworkInfo(sourceURL)
	if err != nil {
		return nil, err
	}
	if index >= uint(len(artwork.Pictures)) {
		return nil, ErrIndexOOB
	}
	return artwork.Pictures[index], nil
}

func (k *Kemono) GetSourceURLRegexp() *regexp.Regexp {
	return kemonoSourceURLRegex
}

func (k *Kemono) GetCommonSourceURL(url string) string {
	return "https://" + kemonoSourceURLRegex.FindString(url)
}

func (k *Kemono) GetFileName(artwork *types.Artwork, picture *types.Picture) string {
	return artwork.Title + "_" + strconv.Itoa(int(picture.Index)) + "_" + filepath.Base(picture.Original)
}

func (k *Kemono) Config() *config.SourceCommonConfig {
	return &config.SourceCommonConfig{
		Enable:   config.Cfg.Source.Kemono.Enable,
		Intervel: -1,
	}
}
