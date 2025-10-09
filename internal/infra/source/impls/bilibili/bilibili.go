package bilibili

import (
	"context"

	"github.com/krau/ManyACG/internal/infra/config/runtimecfg"
	"github.com/krau/ManyACG/internal/infra/source"
	"github.com/krau/ManyACG/internal/model/dto"
	"github.com/krau/ManyACG/internal/shared"

	"github.com/imroc/req/v3"
)

type Bilibili struct {
	cfg       runtimecfg.SourceBilibiliConfig
	reqClient *req.Client
}

// PrettyFileName implements source.ArtworkSource.
func (b *Bilibili) PrettyFileName(artwork shared.ArtworkLike, picture shared.PictureLike) string {
	panic("unimplemented")
}

func init() {
	source.Register(shared.SourceTypeBilibili, func() source.ArtworkSource {
		return &Bilibili{
			cfg:       runtimecfg.Get().Source.Bilibili,
			reqClient: req.C().SetCommonRetryCount(3).ImpersonateChrome(),
		}
	})
}

func (b *Bilibili) FetchNewArtworks(ctx context.Context, limit int) ([]*dto.FetchedArtwork, error) {
	return nil, nil
}

func (b *Bilibili) GetArtworkInfo(ctx context.Context, sourceURL string) (*dto.FetchedArtwork, error) {
	// dynamicID := getDynamicID(sourceURL)
	// if dynamicID == "" {
	// 	return nil, ErrInvalidURL
	// }
	// var err error
	// var desktopResp *BilibiliDesktopDynamicApiResp
	// desktopResp, err = reqDesktopDynamicApiResp(dynamicID)
	// if err == nil {
	// 	var artwork *types.Artwork
	// 	artwork, err = desktopResp.ToArtwork()
	// 	if errors.Is(err, ErrInvalidURL) {
	// 		return nil, err
	// 	}
	// 	if err == nil {
	// 		return artwork, nil
	// 	}
	// }
	// var webResp *BilibiliWebDynamicApiResp
	// webResp, err = reqWebDynamicApiResp(dynamicID)
	// if err == nil {
	// 	var artwork *types.Artwork
	// 	artwork, err = webResp.ToArtwork()
	// 	if err == nil {
	// 		return artwork, nil
	// 	}
	// }
	// return nil, err
	panic("not implemented")
}

func (b *Bilibili) MatchesSourceURL(text string) (string, bool) {
	dynamicID := getDynamicID(text)
	if dynamicID == "" {
		return "", false
	}
	return "https://t.bilibili.com/" + dynamicID, true
}

// func (b *Bilibili) GetFileName(artwork *types.Artwork, picture *types.Picture) string {
// 	dynamicID := getDynamicID(artwork.SourceURL)
// 	return fmt.Sprintf("%s_%d%s", dynamicID, picture.Index, path.Ext(picture.Original))
// }

// func (b *Bilibili) Config() *config.SourceCommonConfig {
// 	return &config.SourceCommonConfig{
// 		Enable:   config.Get().Source.Bilibili.Enable,
// 		Intervel: -1,
// 	}
// }
