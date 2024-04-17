package sources

import (
	"ManyACG-Bot/config"
	"ManyACG-Bot/types"
)

type Source interface {
	FetchNewArtworks(artworkCh chan *types.Artwork, limit int) error
	GetArtworkInfo(sourceURL string) (*types.Artwork, error)
	GetPictureInfo(sourceURL string, index uint) (*types.Picture, error)
	Config() *config.SourceCommonConfig
}
