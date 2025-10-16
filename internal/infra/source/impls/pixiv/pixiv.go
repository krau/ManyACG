package pixiv

import (
	"context"
	"fmt"
	"net/http"

	config "github.com/krau/ManyACG/internal/infra/config/runtimecfg"
	"github.com/krau/ManyACG/internal/infra/source"
	"github.com/krau/ManyACG/internal/model/dto"
	"github.com/krau/ManyACG/internal/shared"
	"github.com/krau/ManyACG/pkg/log"
	"github.com/krau/ManyACG/pkg/strutil"
	"github.com/samber/oops"

	"github.com/imroc/req/v3"
)

type Pixiv struct {
	cfg       config.SourcePixivConfig
	reqClient *req.Client
}

func Init() {
	cfg := config.Get().Source.Pixiv
	if cfg.Disable {
		return
	}

	source.Register(shared.SourceTypePixiv, func() source.ArtworkSource {
		cfg := config.Get().Source.Pixiv
		cookies := make([]*http.Cookie, 0)
		for _, cookie := range cfg.Cookies {
			cookies = append(cookies, &http.Cookie{
				Name:  cookie.Name,
				Value: cookie.Value,
			})
		}
		c := req.C().ImpersonateChrome().SetCommonCookies(cookies...)
		c = c.SetLogger(log.Default()).EnableDebugLog().SetCommonRetryCount(3)
		if config.Get().Source.Proxy != "" {
			c.SetProxyURL(config.Get().Source.Proxy)
		}
		return &Pixiv{
			cfg:       cfg,
			reqClient: c,
		}
	})
}

func (p *Pixiv) FetchNewArtworks(ctx context.Context, limit int) ([]*dto.FetchedArtwork, error) {
	artworks := make([]*dto.FetchedArtwork, 0)
	errs := make([]error, 0)
	for _, url := range p.cfg.RssURLs {
		artworksForURL, err := p.fetchNewArtworksForRSSURL(ctx, url, limit)
		if err != nil {
			errs = append(errs, err)
		}
		artworks = append(artworks, artworksForURL...)
	}
	if len(errs) > 0 {
		return artworks, fmt.Errorf("fetching pixiv encountered %d errors: %v", len(errs), errs)
	}
	return artworks, nil
}

func (p *Pixiv) GetArtworkInfo(ctx context.Context, sourceURL string) (*dto.FetchedArtwork, error) {
	ajaxResp, err := reqAjaxResp(ctx, sourceURL, p.reqClient)
	if err != nil {
		return nil, err
	}
	if ajaxResp.Err {
		return nil, oops.Wrapf(err, "pixiv ajax response error: %s", ajaxResp.Message)
	}
	return ajaxResp.ToArtwork(ctx, p.reqClient, p.cfg.ImgProxy)
}

func (p *Pixiv) MatchesSourceURL(text string) (string, bool) {
	pid := getPid(text)
	if pid == "" {
		return "", false
	}
	return "https://www.pixiv.net/artworks/" + pid, true
}

func (p *Pixiv) PrettyFileName(artwork shared.ArtworkLike, picture shared.PictureLike) string {
	pid := getPid(artwork.GetSourceURL())
	ext, _ := strutil.GetFileExtFromURL(picture.GetOriginal())
	if pid != "" {
		return fmt.Sprintf("pixiv_%s_%d%s", pid, picture.GetIndex(), ext)
	}
	return fmt.Sprintf("pixiv_%s%s", strutil.MD5Hash(picture.GetOriginal()), ext)
}
