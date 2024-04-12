package pixiv

import (
	"ManyACG-Bot/config"
	"ManyACG-Bot/model"
	"sync"
)

type Pixiv struct{}

func (p *Pixiv) GetNewArtworks(limit int) ([]model.Artwork, error) {
	artworks := make([]model.Artwork, 0)

	var wg sync.WaitGroup

	artworkChan := make(chan *model.Artwork, len(config.Cfg.Source.Pixiv.URLs)*limit)

	for _, url := range config.Cfg.Source.Pixiv.URLs {
		wg.Add(1)
		go getNewArtworksForURL(url, limit, &wg, artworkChan)
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

func (p *Pixiv) Config() *config.SourceCommonConfig {
	return &config.SourceCommonConfig{
		Enable:   config.Cfg.Source.Pixiv.Enable,
		Intervel: config.Cfg.Source.Pixiv.Intervel,
	}
}
