package sources

import (
	"regexp"

	"github.com/krau/ManyACG/config"
	"github.com/krau/ManyACG/types"
)

type Source interface {
	Init(service types.Service)
	FetchNewArtworksWithCh(artworkCh chan *types.Artwork, limit int) error
	FetchNewArtworks(limit int) ([]*types.Artwork, error)
	GetArtworkInfo(sourceURL string) (*types.Artwork, error)
	GetPictureInfo(sourceURL string, index uint) (*types.Picture, error)
	GetSourceURLRegexp() *regexp.Regexp
	// CommonSourceURl should has prefix "https://"
	GetCommonSourceURL(url string) string
	// FileName 返回图片的用于存储和传输的文件名
	GetFileName(artwork *types.Artwork, picture *types.Picture) string
	Config() *config.SourceCommonConfig
}
