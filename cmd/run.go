package cmd

import (
	"ManyACG-Bot/dao"
	. "ManyACG-Bot/logger"
	"ManyACG-Bot/model"
	"ManyACG-Bot/sources"
	"ManyACG-Bot/telegram"
	"time"
)

func Run() {
	Logger.Info("Start running")
	artworkCh := make(chan model.Artwork, 30)

	for name, source := range sources.Sources {
		Logger.Infof("Start fetching from %s", name)
		go func(source sources.Source, limit int, ch chan model.Artwork, intervel uint) {
			ticker := time.NewTicker(time.Duration(intervel) * time.Minute)
			for {
				artworks, err := source.GetNewArtworks(limit)
				if err != nil {
					Logger.Errorf("Error when fetching from %s: %s", name, err)
				}
				if len(artworks) > 0 {
					Logger.Infof("Fetched %d artworks from %s", len(artworks), name)
					for _, artwork := range artworks {
						ch <- artwork
					}
				}
				<-ticker.C
			}
		}(source, 30, artworkCh, source.Config().Intervel)
	}

	for artwork := range artworkCh {
		artworkDB, _ := dao.GetArtworkByURL(artwork.SourceURL)
		if artworkDB.ID != 0 {
			Logger.Infof("Artwork %s already exists", artwork.Title)
			continue
			// TODO: Update artwork
		}
		Logger.Infof("Posting artwork %s", artwork.Title)
		_, err := telegram.PostArtwork(telegram.Bot, &artwork)
		if err != nil {
			Logger.Errorf("Error when posting artwork %s: %s", artwork.Title, err)
			continue
		}

		// TODO: Download pictures and save to database

	}
}
