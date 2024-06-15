package danbooru

import (
	"ManyACG/common"
	"ManyACG/config"
	. "ManyACG/logger"
	"ManyACG/types"
	"encoding/json"
	"errors"
	"path/filepath"
	"regexp"
	"strings"
)

type Danbooru struct{}

func (d *Danbooru) Init() {
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
	resp, err := common.Client.R().Get(sourceURL + ".json")
	if err != nil {
		return nil, err
	}
	var danbooruResp DanbooruSuccessJsonResp
	if err := json.Unmarshal(resp.Bytes(), &danbooruResp); err != nil {
		return nil, err
	}
	if danbooruResp.ID == 0 {
		var danbooruFailResp DanbooruFailJsonResp
		if err := json.Unmarshal(resp.Bytes(), &danbooruFailResp); err != nil {
			return nil, err
		}
		return nil, errors.New(danbooruFailResp.Message)
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
