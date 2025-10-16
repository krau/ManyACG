package bilibili

import (
	"context"
	"errors"
	"fmt"

	"github.com/krau/ManyACG/internal/infra/config/runtimecfg"
	"github.com/krau/ManyACG/internal/infra/source"
	"github.com/krau/ManyACG/internal/model/dto"
	"github.com/krau/ManyACG/internal/shared"
	"github.com/krau/ManyACG/pkg/strutil"

	"github.com/imroc/req/v3"
)

type Bilibili struct {
	cfg       runtimecfg.SourceBilibiliConfig
	reqClient *req.Client
}

func Init() {
	cfg := runtimecfg.Get().Source.Bilibili
	if cfg.Disable {
		return
	}
	client := req.C().SetCommonRetryCount(5).ImpersonateChrome()
	if runtimecfg.Get().Source.Proxy != "" {
		client.SetProxyURL(runtimecfg.Get().Source.Proxy)
	}
	source.Register(shared.SourceTypeBilibili, func() source.ArtworkSource {
		return &Bilibili{
			cfg:       runtimecfg.Get().Source.Bilibili,
			reqClient: client,
		}
	})
}

func (b *Bilibili) FetchNewArtworks(ctx context.Context, limit int) ([]*dto.FetchedArtwork, error) {
	return nil, nil
}

func (b *Bilibili) GetArtworkInfo(ctx context.Context, sourceURL string) (*dto.FetchedArtwork, error) {
	dynamicID := getDynamicID(sourceURL)
	if dynamicID == "" {
		return nil, ErrInvalidURL
	}
	var err error
	var desktopResp *BilibiliDesktopDynamicApiResp
	desktopResp, err = b.reqDesktopDynamicApiResp(ctx, dynamicID)
	if err == nil {
		var artwork *dto.FetchedArtwork
		artwork, err = desktopResp.ToArtwork()
		if errors.Is(err, ErrInvalidURL) {
			return nil, err
		}
		if err == nil {
			return artwork, nil
		}
	}
	var webResp *BilibiliWebDynamicApiResp
	webResp, err = b.reqWebDynamicApiResp(ctx, dynamicID)
	if err == nil {
		var artwork *dto.FetchedArtwork
		artwork, err = webResp.ToArtwork()
		if err == nil {
			return artwork, nil
		}
	}
	return nil, err
}

func (b *Bilibili) MatchesSourceURL(text string) (string, bool) {
	dynamicID := getDynamicID(text)
	if dynamicID == "" {
		return "", false
	}
	return "https://t.bilibili.com/" + dynamicID, true
}

// PrettyFileName implements source.ArtworkSource.
func (b *Bilibili) PrettyFileName(artwork shared.ArtworkLike, picture shared.PictureLike) string {
	id := getDynamicID(artwork.GetSourceURL())
	ext, _ := strutil.GetFileExtFromURL(picture.GetOriginal())
	if id == "" {
		return fmt.Sprintf("bilibili_%s%s", strutil.MD5Hash(picture.GetOriginal()), ext)
	}
	return fmt.Sprintf("bilibili_%s_%d%s", id, picture.GetIndex(), ext)
}
