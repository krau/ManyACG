package nhentai

import (
	"context"

	"github.com/imroc/req/v3"
	"github.com/krau/ManyACG/internal/infra/config"
	"github.com/krau/ManyACG/internal/infra/source"
	"github.com/krau/ManyACG/internal/shared"
)

type Nhentai struct {
	cfg       config.SourceNhentaiConfig
	reqClient *req.Client
}

func init() {
	source.Register(shared.SourceTypeNhentai, func() source.ArtworkSource {
		return &Nhentai{
			cfg:       config.Get().Source.Nhentai,
			reqClient: req.C().ImpersonateChrome().SetCommonRetryCount(3),
		}
	})
}

func (n *Nhentai) FetchNewArtworks(ctx context.Context, limit int) ([]*source.FetchedArtwork, error) {
	return nil, nil
}

func (n *Nhentai) GetArtworkInfo(ctx context.Context, sourceURL string) (*source.FetchedArtwork, error) {
	// galleryID := GetGalleryID(sourceURL)
	// if galleryID == "" {
	// 	return nil, ErrorInvalidNhentaiURL
	// }
	// common.Logger.Tracef("request artwork info: %s", sourceURLPrefix+galleryID)
	// return n.crawlGallery(galleryID)
	panic("not implemented")
}

func (n *Nhentai) MatchesSourceURL(text string) (string, bool) {
	galleryID := GetGalleryID(text)
	if galleryID == "" {
		return "", false
	}
	return sourceURLPrefix + galleryID, true
}

// func (n *Nhentai) GetFileName(artwork *types.Artwork, picture *types.Picture) string {
// 	galleryID := GetGalleryID(artwork.SourceURL)
// 	if galleryID == "" {
// 		galleryID = picture.ID
// 	}
// 	if galleryID == "" {
// 		galleryID = common.MD5Hash(picture.Original)
// 	}
// 	return "nhentai_" + galleryID + "_" + strconv.Itoa(int(picture.Index)) + path.Ext(picture.Original)
// }

// func (n *Nhentai) Config() *config.SourceCommonConfig {
// 	return &config.SourceCommonConfig{
// 		Enable:   config.Cfg.Source.Nhentai.Enable,
// 		Intervel: -1,
// 	}
// }
