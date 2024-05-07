package sources

import (
	"ManyACG-Bot/config"
	"ManyACG-Bot/types"
	"regexp"
)

type Source interface {
	Init()
	FetchNewArtworksWithCh(artworkCh chan *types.Artwork, limit int) error
	FetchNewArtworks(limit int) ([]*types.Artwork, error)
	GetArtworkInfo(sourceURL string) (*types.Artwork, error)
	GetPictureInfo(sourceURL string, index uint) (*types.Picture, error)
	GetSourceURLRegexp() *regexp.Regexp
	// CommonSourceURl should has prefix "https://"
	GetCommonSourceURL(url string) string
	Config() *config.SourceCommonConfig
}
