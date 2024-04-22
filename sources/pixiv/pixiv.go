package pixiv

import (
	"ManyACG-Bot/config"
	"ManyACG-Bot/types"
	"errors"
	"fmt"
	"strings"
)

type Pixiv struct{}

func (p *Pixiv) FetchNewArtworksWithCh(artworkCh chan *types.Artwork, limit int) error {
	errs := make([]error, 0)

	for _, url := range config.Cfg.Source.Pixiv.URLs {
		err := fetchNewArtworksForRSSURLWithCh(url, limit, artworkCh)
		if err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("encountered %d errors: %v", len(errs), errs)
	}

	return nil
}

func (p *Pixiv) FetchNewArtworks(limit int) ([]*types.Artwork, error) {
	artworks := make([]*types.Artwork, 0)
	errs := make([]error, 0)
	for _, url := range config.Cfg.Source.Pixiv.URLs {
		artworksForURL, err := fetchNewArtworksForRSSURL(url, limit)
		if err != nil {
			errs = append(errs, err)
		}
		artworks = append(artworks, artworksForURL...)
	}
	if len(errs) > 0 {
		return artworks, fmt.Errorf("encountered %d errors: %v", len(errs), errs)
	}
	return artworks, nil
}

func (p *Pixiv) GetArtworkInfo(sourceURL string) (*types.Artwork, error) {
	ajaxResp, err := reqAjaxResp(sourceURL)
	if err != nil {
		return nil, err
	}
	if ajaxResp.Err {
		return nil, errors.New(ajaxResp.Message)
	}
	return ajaxResp.ToArtwork()
}

func (p *Pixiv) GetPictureInfo(sourceURL string, index uint) (*types.Picture, error) {
	resp, err := reqIllustPages(sourceURL)
	if err != nil {
		return nil, err
	}
	if resp.Err {
		return nil, errors.New(resp.Message)
	}
	return &types.Picture{
		Index:        index,
		Thumbnail:    strings.Replace(resp.Body[index].Urls.Small, "i.pximg.net", config.Cfg.Source.Pixiv.Proxy, 1),
		Original:     strings.Replace(resp.Body[index].Urls.Original, "i.pximg.net", config.Cfg.Source.Pixiv.Proxy, 1),
		Width:        uint(resp.Body[index].Width),
		Height:       uint(resp.Body[index].Height),
		TelegramInfo: &types.TelegramInfo{},
	}, nil
}

func (p *Pixiv) Config() *config.SourceCommonConfig {
	return &config.SourceCommonConfig{
		Enable:   config.Cfg.Source.Pixiv.Enable,
		Intervel: config.Cfg.Source.Pixiv.Intervel,
	}
}
