package sources

import (
	"ManyACG-Bot/config"
	"ManyACG-Bot/model"
)

type Source interface {
	GetNewArtworks(limit int) ([]model.Artwork, error)
	// GetArtworkByURL(url string) (model.Artwork, error)
	Config() *config.SourceCommonConfig
}
