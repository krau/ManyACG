package kemono

import (
	"ManyACG/config"
	. "ManyACG/logger"
	"ManyACG/types"
	"encoding/json"
	"errors"
	"net/http"
	"path/filepath"
	"regexp"
	"strconv"
)

type Kemono struct{}

func (k *Kemono) Init() {
	if config.Cfg.Source.Kemono.Session != "" {
		reqClient.SetCommonCookies(&http.Cookie{
			Name:  "session",
			Value: config.Cfg.Source.Kemono.Session,
		})
	}
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
	Logger.Tracef("request artwork info: %s", apiURL)
	resp, err := reqClient.R().Get(apiURL)
	if err != nil {
		return nil, err
	}
	var kemonoResp KemonoPostResp
	if err := json.Unmarshal(resp.Bytes(), &kemonoResp); err != nil {
		return nil, err
	}
	if kemonoResp.Error != "" {
		return nil, errors.New(kemonoResp.Error)
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
	return "https://" + url
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
