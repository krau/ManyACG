package pixiv

import (
	"ManyACG-Bot/config"
	"ManyACG-Bot/types"
	"errors"
	"sync"
)

type Pixiv struct{}

func (p *Pixiv) FetchNewArtworks(limit int) ([]types.Artwork, error) {
	artworks := make([]types.Artwork, 0)

	var wg sync.WaitGroup

	artworkChan := make(chan *types.Artwork, len(config.Cfg.Source.Pixiv.URLs)*limit)

	for _, url := range config.Cfg.Source.Pixiv.URLs {
		wg.Add(1)
		go fetchNewArtworksForRSSURL(url, limit, &wg, artworkChan)
	}

	go func() {
		wg.Wait()
		close(artworkChan)
	}()

	for artwork := range artworkChan {
		artworks = append(artworks, *artwork)
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
	return nil, nil
}

func (p *Pixiv) Config() *config.SourceCommonConfig {
	return &config.SourceCommonConfig{
		Enable:   config.Cfg.Source.Pixiv.Enable,
		Intervel: config.Cfg.Source.Pixiv.Intervel,
	}
}
