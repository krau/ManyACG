package pixiv

import (
	"context"
	"fmt"
	"net/http"

	config "github.com/krau/ManyACG/internal/infra/config/runtimecfg"
	"github.com/krau/ManyACG/internal/infra/kvstor"
	"github.com/krau/ManyACG/internal/infra/source"
	"github.com/krau/ManyACG/internal/model/dto"
	"github.com/krau/ManyACG/internal/shared"
	"github.com/krau/ManyACG/pkg/log"
	"github.com/krau/ManyACG/pkg/strutil"
	"github.com/samber/oops"

	"github.com/imroc/req/v3"
)

type Namer struct{}

type Pixiv struct {
	Namer
	reqClient *req.Client
	cfg       config.SourcePixivConfig
}

// Init registers the Pixiv detail source. A non-nil store additionally
// registers an RSS poller that caches fetched artworks in the store.
func Init(store kvstor.KVStore) {
	cfg := config.Get().Source.Pixiv
	if cfg.Disable {
		return
	}
	client := newRequestClient(cfg)
	source.Register(shared.SourceTypePixiv, func() source.ArtworkSource {
		return &Pixiv{
			cfg:       cfg,
			reqClient: client,
		}
	})
	if store != nil {
		source.RegisterPoller(shared.SourceTypePixiv, &Poller{
			cfg:       cfg,
			reqClient: client,
			kv:        store,
		})
	}
}

func newRequestClient(cfg config.SourcePixivConfig) *req.Client {
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
	return c
}

func (p *Pixiv) GetArtworkInfo(ctx context.Context, sourceURL string) (*dto.FetchedArtwork, error) {
	ajaxResp, err := reqAjaxResp(ctx, sourceURL, p.reqClient)
	if err != nil {
		return nil, err
	}
	if ajaxResp.Err {
		return nil, oops.Errorf("pixiv ajax response error: %s", ajaxResp.Message)
	}
	return ajaxResp.ToArtwork(ctx, p.reqClient, p.cfg.ImgProxy)
}

func (Namer) MatchesSourceURL(text string) (string, bool) {
	pid := getPid(text)
	if pid == "" {
		return "", false
	}
	return "https://www.pixiv.net/artworks/" + pid, true
}

func (Namer) PrettyFileName(artwork shared.ArtworkLike, picture shared.PictureLike) string {
	pid := getPid(artwork.GetSourceURL())
	ext, _ := strutil.GetFileExtFromURL(picture.GetOriginal())
	if pid != "" {
		return fmt.Sprintf("pixiv_%s_%d%s", pid, picture.GetIndex(), ext)
	}
	return fmt.Sprintf("pixiv_%s%s", strutil.MD5Hash(picture.GetOriginal()), ext)
}
