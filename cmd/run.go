package cmd

import (
	. "ManyACG-Bot/logger"
	"ManyACG-Bot/model"
	"ManyACG-Bot/sources"
	"time"
)

func Run() {
	Logger.Info("Start running")
	artworkCh := make(chan []model.Artwork, 30)

	for name, source := range sources.Sources {
		Logger.Infof("Start fetching from %s", name)
		go func(source sources.Source, limit int, ch chan []model.Artwork, intervel uint) {
			ticker := time.NewTicker(time.Duration(intervel) * time.Minute)
			for {
				artworks, err := source.GetNewArtworks(limit)
				if err != nil {
					Logger.Errorf("Error when fetching from %s: %s", name, err)
				}
				if len(artworks) > 0 {
					Logger.Infof("Fetched %d artworks from %s", len(artworks), name)
					ch <- artworks
				}
				<-ticker.C
			}
		}(source, 30, artworkCh, source.Config().Intervel)
	}
}
