package pixiv

import (
	"ManyACG-Bot/config"
	"ManyACG-Bot/types"
	"errors"
	"fmt"
	"strings"
	"sync"
)

type Pixiv struct{}

func (p *Pixiv) FetchNewArtworks(artworkCh chan *types.Artwork, limit int) error {
	var wg sync.WaitGroup
	errs := make(chan error, len(config.Cfg.Source.Pixiv.URLs))

	for _, url := range config.Cfg.Source.Pixiv.URLs {
		wg.Add(1)
		go func(url string) {
			defer wg.Done()
			err := fetchNewArtworksForRSSURL(url, limit, artworkCh)
			if err != nil {
				errs <- err
			}
		}(url)
	}

	go func() {
		wg.Wait()
		close(errs)
	}()

	var totalErr []error
	for err := range errs {
		totalErr = append(totalErr, err)
	}

	if len(totalErr) > 0 {
		return fmt.Errorf("encountered %d errors: %v", len(totalErr), totalErr)
	}

	return nil
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
