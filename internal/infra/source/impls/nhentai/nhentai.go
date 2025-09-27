package nhentai

import (
	"path"
	"regexp"
	"strconv"

	"github.com/imroc/req/v3"
	"github.com/krau/ManyACG/common"
	"github.com/krau/ManyACG/config"
	sourceCommon "github.com/krau/ManyACG/internal/infra/source/common"
	"github.com/krau/ManyACG/types"
)

type Nhentai struct{}

var reqClient *req.Client

func init() {
	sourceCommon.RegisterSource(types.SourceTypeNhentai, new(Nhentai))
}

func (n *Nhentai) Init(_ types.Service) {
	reqClient = req.C().ImpersonateChrome().SetCommonRetryCount(2)
	if config.Cfg.Source.Proxy != "" {
		reqClient.SetProxyURL(config.Cfg.Source.Proxy)
	}
}

func (n *Nhentai) FetchNewArtworksWithCh(artworkCh chan *types.Artwork, limit int) error {
	return nil
}

func (n *Nhentai) FetchNewArtworks(limit int) ([]*types.Artwork, error) {
	return nil, nil
}

func (n *Nhentai) GetArtworkInfo(sourceURL string) (*types.Artwork, error) {
	galleryID := GetGalleryID(sourceURL)
	if galleryID == "" {
		return nil, ErrorInvalidNhentaiURL
	}
	common.Logger.Tracef("request artwork info: %s", sourceURLPrefix+galleryID)
	return n.crawlGallery(galleryID)
}

func (n *Nhentai) GetPictureInfo(sourceURL string, i uint) (*types.Picture, error) {
	// TODO: refactor this
	galleryID := GetGalleryID(sourceURL)
	if galleryID == "" {
		return nil, ErrorInvalidNhentaiURL
	}
	common.Logger.Tracef("request picture info: %s", sourceURLPrefix+galleryID)
	artwork, err := n.crawlGallery(galleryID)
	if err != nil {
		return nil, err
	}
	if i >= uint(len(artwork.Pictures)) {
		return nil, ErrorInvalidNhentaiURL
	}
	return artwork.Pictures[i], nil
}

func (n *Nhentai) GetSourceURLRegexp() *regexp.Regexp {
	return nhentaiSourceURLRegexp
}

func (n *Nhentai) GetCommonSourceURL(url string) string {
	galleryID := GetGalleryID(url)
	if galleryID == "" {
		return ""
	}
	return sourceURLPrefix + galleryID
}

func (n *Nhentai) GetFileName(artwork *types.Artwork, picture *types.Picture) string {
	galleryID := GetGalleryID(artwork.SourceURL)
	if galleryID == "" {
		galleryID = picture.ID
	}
	if galleryID == "" {
		galleryID = common.MD5Hash(picture.Original)
	}
	return "nhentai_" + galleryID + "_" + strconv.Itoa(int(picture.Index)) + path.Ext(picture.Original)
}

func (n *Nhentai) Config() *config.SourceCommonConfig {
	return &config.SourceCommonConfig{
		Enable:   config.Cfg.Source.Nhentai.Enable,
		Intervel: -1,
	}
}
