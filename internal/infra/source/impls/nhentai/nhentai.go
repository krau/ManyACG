package nhentai

import (
	"context"
	"fmt"

	"github.com/imroc/req/v3"
	config "github.com/krau/ManyACG/internal/infra/config/runtimecfg"
	"github.com/krau/ManyACG/internal/infra/source"
	"github.com/krau/ManyACG/internal/model/dto"
	"github.com/krau/ManyACG/internal/shared"
	"github.com/krau/ManyACG/pkg/strutil"
)

type Nhentai struct {
	cfg       config.SourceNhentaiConfig
	reqClient *req.Client
}

func Init() {
	cfg := config.Get().Source.Nhentai
	if cfg.Disable {
		return
	}
	source.Register(shared.SourceTypeNhentai, func() source.ArtworkSource {
		return &Nhentai{
			cfg:       config.Get().Source.Nhentai,
			reqClient: req.C().ImpersonateChrome().SetCommonRetryCount(3),
		}
	})
}

func (n *Nhentai) FetchNewArtworks(ctx context.Context, limit int) ([]*dto.FetchedArtwork, error) {
	return nil, nil
}

func (n *Nhentai) GetArtworkInfo(ctx context.Context, sourceURL string) (*dto.FetchedArtwork, error) {
	galleryID := GetGalleryID(sourceURL)
	if galleryID == "" {
		return nil, ErrorInvalidNhentaiURL
	}
	return n.crawlGallery(ctx, galleryID)
}

func (n *Nhentai) MatchesSourceURL(text string) (string, bool) {
	galleryID := GetGalleryID(text)
	if galleryID == "" {
		return "", false
	}
	return sourceURLPrefix + galleryID, true
}

// PrettyFileName implements source.ArtworkSource.
func (n *Nhentai) PrettyFileName(artwork shared.ArtworkLike, picture shared.PictureLike) string {
	galleryID := GetGalleryID(artwork.GetSourceURL())
	ext, _ := strutil.GetFileExtFromURL(picture.GetOriginal())
	if galleryID == "" {
		return fmt.Sprintf("nhentai_%s%s", strutil.MD5Hash(picture.GetOriginal()), ext)
	}
	return fmt.Sprintf("nhentai_%s_%d%s", galleryID, picture.GetIndex(), ext)
}
