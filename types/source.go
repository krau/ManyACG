package types

import (
	"regexp"

	"github.com/krau/ManyACG/config"
)

type SourceType string

const (
	SourceTypePixiv    SourceType = "pixiv"
	SourceTypeTwitter  SourceType = "twitter"
	SourceTypeBilibili SourceType = "bilibili"
	SourceTypeDanbooru SourceType = "danbooru"
	SourceTypeKemono   SourceType = "kemono"
	SourceTypeYandere  SourceType = "yandere"
)

var SourceTypes []SourceType = []SourceType{
	SourceTypePixiv,
	SourceTypeTwitter,
	SourceTypeBilibili,
	SourceTypeDanbooru,
	SourceTypeKemono,
	SourceTypeYandere,
}

type Source interface {
	Init(service Service)
	FetchNewArtworksWithCh(artworkCh chan *Artwork, limit int) error
	FetchNewArtworks(limit int) ([]*Artwork, error)
	GetArtworkInfo(sourceURL string) (*Artwork, error)
	GetPictureInfo(sourceURL string, index uint) (*Picture, error)
	GetSourceURLRegexp() *regexp.Regexp
	// CommonSourceURl should has prefix "https://"
	GetCommonSourceURL(url string) string
	// FileName 返回图片的用于存储和传输的文件名
	GetFileName(artwork *Artwork, picture *Picture) string
	Config() *config.SourceCommonConfig
}
