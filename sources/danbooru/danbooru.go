package danbooru

import (
	"encoding/json"
	"errors"
	"fmt"
	"path"
	"regexp"

	"github.com/imroc/req/v3"
	"github.com/krau/ManyACG/common"
	"github.com/krau/ManyACG/config"
	sourceCommon "github.com/krau/ManyACG/sources/common"

	"github.com/krau/ManyACG/types"
)

type Danbooru struct{}

var reqClient *req.Client

func init() {
	sourceCommon.RegisterSource(types.SourceTypeDanbooru, new(Danbooru))
}

func (d *Danbooru) Init(_ types.Service) {
	reqClient = req.C().SetCommonRetryCount(2)
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
	postID := GetPostID(sourceURL)
	if postID == "" {
		return nil, ErrInvalidDanbooruPostURL
	}
	sourceURL = "https://danbooru.donmai.us/posts/" + postID
	common.Logger.Tracef("request artwork info: %s", sourceURL)
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
	postID := GetPostID(url)
	if postID == "" {
		return ""
	}
	return "https://danbooru.donmai.us/posts/" + postID
}

func (d *Danbooru) GetFileName(artwork *types.Artwork, picture *types.Picture) string {
	idStr := GetPostID(artwork.SourceURL)
	return fmt.Sprintf("%s_%d%s", idStr, picture.Index, path.Ext(picture.Original))
}

func (d *Danbooru) Config() *config.SourceCommonConfig {
	return &config.SourceCommonConfig{
		Enable:   config.Cfg.Source.Danbooru.Enable,
		Intervel: -1,
	}
}
