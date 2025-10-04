package twitter

import (
	"context"
	"net/url"

	"github.com/imroc/req/v3"
	"github.com/krau/ManyACG/internal/infra/config/runtimecfg"
	"github.com/krau/ManyACG/internal/infra/source"
	"github.com/krau/ManyACG/internal/model/dto"
	"github.com/krau/ManyACG/internal/shared"
)

type Twitter struct {
	cfg       runtimecfg.SourceTwitterConfig
	reqClient *req.Client
}

func init() {
	source.Register(shared.SourceTypeTwitter, func() source.ArtworkSource {
		return &Twitter{
			cfg:       runtimecfg.Get().Source.Twitter,
			reqClient: req.C(),
		}
	})
}

func (t *Twitter) FetchNewArtworks(ctx context.Context, limit int) ([]*dto.FetchedArtwork, error) {
	// artworks := make([]*types.Artwork, 0)
	// errs := make([]error, 0)
	// for _, url := range config.Get().Source.Twitter.URLs {
	// 	artworksForURL, err := t.fetchRssURL(url, limit)
	// 	if err != nil {
	// 		errs = append(errs, err)
	// 	}
	// 	artworks = append(artworks, artworksForURL...)
	// }
	// if len(errs) > 0 {
	// 	return nil, fmt.Errorf("fetching twitter encountered %d errors: %v", len(errs), errs)
	// }
	// return artworks, nil
	panic("not implemented")
}

func (t *Twitter) GetArtworkInfo(ctx context.Context, sourceURL string) (*dto.FetchedArtwork, error) {
	// tweetID := getTweetID(sourceURL)
	// if tweetID == "" {
	// 	return nil, ErrInvalidURL
	// }
	// fxTwitterApiURL := fmt.Sprintf("https://api.%s/_/status/%s", config.Get().Source.Twitter.FxTwitterDomain, tweetID)
	// resp, err := reqApiResp(fxTwitterApiURL)
	// if err != nil {
	// 	return nil, err
	// }
	// return resp.ToArtwork()
	panic("not implemented")
}

func (t *Twitter) MatchesSourceURL(text string) (string, bool) {
	matches := twitterSourceURLRegexp.FindString(text)
	if len(matches) == 0 {
		return "", false
	}
	tweet := getTweetPath(text)
	if tweet == "" {
		return "", false
	}
	commonUrl, err := url.JoinPath("https://x.com", tweet)
	if err != nil {
		return "", false
	}
	return commonUrl, true
}

// func (t *Twitter) GetFileName(artwork *types.Artwork, picture *types.Picture) string {
// 	original := picture.Original
// 	urlSplit := strings.Split(picture.Original, "?")
// 	if len(urlSplit) > 1 {
// 		original = strings.Join(urlSplit[:len(urlSplit)-1], "?")
// 	}
// 	tweetID := strings.Split(artwork.SourceURL, "/")[len(strings.Split(artwork.SourceURL, "/"))-1]
// 	return tweetID + "_" + strconv.Itoa(int(picture.Index)) + path.Ext(original)
// }

// func (t *Twitter) Config() *config.SourceCommonConfig {
// 	return &config.SourceCommonConfig{
// 		Enable:   config.Get().Source.Twitter.Enable,
// 		Intervel: config.Get().Source.Twitter.Intervel,
// 	}

// }
