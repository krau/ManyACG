package pixiv

import (
	"context"
	"net/http"

	config "github.com/krau/ManyACG/internal/infra/config/runtimecfg"
	"github.com/krau/ManyACG/internal/infra/source"
	"github.com/krau/ManyACG/internal/model/dto"
	"github.com/krau/ManyACG/internal/shared"

	"github.com/imroc/req/v3"
)

type Pixiv struct {
	cfg       config.SourcePixivConfig
	reqClient *req.Client
}

func init() {
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
		return &Pixiv{
			cfg:       cfg,
			reqClient: c,
		}
	})
}

func (p *Pixiv) FetchNewArtworks(ctx context.Context, limit int) ([]*dto.FetchedArtwork, error) {
	// artworks := make([]*types.Artwork, 0)
	// errs := make([]error, 0)
	// for _, url := range config.Get().Source.Pixiv.URLs {
	// 	artworksForURL, err := fetchNewArtworksForRSSURL(url, limit)
	// 	if err != nil {
	// 		errs = append(errs, err)
	// 	}
	// 	artworks = append(artworks, artworksForURL...)
	// }
	// if len(errs) > 0 {
	// 	return artworks, fmt.Errorf("fetching pixiv encountered %d errors: %v", len(errs), errs)
	// }
	// return artworks, nil
	panic("not implemented")
}

func (p *Pixiv) GetArtworkInfo(ctx context.Context, sourceURL string) (*dto.FetchedArtwork, error) {
	// ajaxResp, err := reqAjaxResp(sourceURL)
	// if err != nil {
	// 	return nil, err
	// }
	// if ajaxResp.Err {
	// 	return nil, errors.New(ajaxResp.Message)
	// }
	// return ajaxResp.ToArtwork()
	panic("not implemented")
}

func (p *Pixiv) MatchesSourceURL(text string) (string, bool) {
	pid := GetPid(text)
	if pid == "" {
		return "", false
	}
	return "https://www.pixiv.net/artworks/" + pid, true
}

// func (p *Pixiv) GetFileName(artwork *types.Artwork, picture *types.Picture) string {
// 	return artwork.Title + "_" + path.Base(picture.Original)
// }

// func (p *Pixiv) Config() *config.SourceCommonConfig {
// 	return &config.SourceCommonConfig{
// 		Enable:   config.Get().Source.Pixiv.Enable,
// 		Intervel: config.Get().Source.Pixiv.Intervel,
// 	}
// }
