package danbooru

import (
	"encoding/json"
	"errors"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/imroc/req/v3"
	"github.com/krau/ManyACG/config"
	. "github.com/krau/ManyACG/logger"
	"github.com/krau/ManyACG/types"
)

type Danbooru struct{}

var reqClient *req.Client

func (d *Danbooru) Init() {
	reqClient = req.C().ImpersonateChrome().SetCommonRetryCount(2)
	if config.Cfg.Source.Proxy != "" {
		reqClient.SetProxyURL(config.Cfg.Source.Proxy)
	}
}

func (d *Danbooru) FetchNewArtworksWithCh(artworkCh chan *types.Artwork, limit int) error {
	return nil
}

func (d *Danbooru) FetchNewArtworks(limit int) ([]*types.Artwork, error) {
	return nil, nil
}

func (d *Danbooru) GetArtworkInfo(sourceURL string) (*types.Artwork, error) {
	danbooruPostURL := danbooruSourceURLRegexp.FindString(sourceURL)
	if danbooruPostURL == "" {
		return nil, ErrInvalidDanbooruPostURL
	}
	sourceURL = "https://" + danbooruPostURL
	Logger.Tracef("request artwork info: %s", sourceURL)
	resp, err := reqClient.R().Get(sourceURL + ".json")
	if err != nil {
		return nil, err
	}
	var danbooruResp DanbooruJsonResp
	if err := json.Unmarshal(resp.Bytes(), &danbooruResp); err != nil {
		return nil, err
	}
	if danbooruResp.Error != "" {
		return nil, errors.New(danbooruResp.Message)
	}
	return danbooruResp.ToArtwork(), nil
}

func (d *Danbooru) GetPictureInfo(sourceURL string, _ uint) (*types.Picture, error) {
	artwork, err := d.GetArtworkInfo(sourceURL)
	if err != nil {
		return nil, err
	}
	return artwork.Pictures[0], nil
}

func (d *Danbooru) GetSourceURLRegexp() *regexp.Regexp {
	return danbooruSourceURLRegexp
}

func (d *Danbooru) GetCommonSourceURL(url string) string {
	danbooruPostURL := danbooruSourceURLRegexp.FindString(url)
	if danbooruPostURL == "" {
		return ""
	}
	return "https://" + danbooruPostURL
}

func (d *Danbooru) GetFileName(artwork *types.Artwork, picture *types.Picture) string {
	idStr := strings.Split(artwork.Title, "/")[1]
	return idStr + filepath.Ext(picture.Original)
}

func (d *Danbooru) Config() *config.SourceCommonConfig {
	return &config.SourceCommonConfig{
		Enable:   config.Cfg.Source.Danbooru.Enable,
		Intervel: -1,
	}
}
