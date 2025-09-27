package twitter

import (
	"fmt"
	"net/url"
	"path"
	"regexp"
	"strconv"
	"strings"

	"github.com/imroc/req/v3"
	"github.com/krau/ManyACG/common"
	"github.com/krau/ManyACG/config"
	sourceCommon "github.com/krau/ManyACG/internal/infra/source/common"
	"github.com/krau/ManyACG/types"
)

type Twitter struct{}

var service types.Service

func init() {
	sourceCommon.RegisterSource(types.SourceTypeTwitter, new(Twitter))
}

func (t *Twitter) Init(s types.Service) {
	reqClient = req.C().ImpersonateChrome().SetCommonRetryCount(3)
	if config.Cfg.Source.Proxy != "" {
		reqClient.SetProxyURL(config.Cfg.Source.Proxy)
	}
	service = s
}

func (t *Twitter) FetchNewArtworksWithCh(artworkCh chan *types.Artwork, limit int) error {
	errs := make([]error, 0)
	for _, url := range config.Cfg.Source.Twitter.URLs {
		err := t.fetchRssURLWithCh(url, limit, artworkCh)
		if err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("fetching twitter encountered %d errors: %v", len(errs), errs)
	}
	return nil
}

func (t *Twitter) FetchNewArtworks(limit int) ([]*types.Artwork, error) {
	artworks := make([]*types.Artwork, 0)
	errs := make([]error, 0)
	for _, url := range config.Cfg.Source.Twitter.URLs {
		artworksForURL, err := t.fetchRssURL(url, limit)
		if err != nil {
			errs = append(errs, err)
		}
		artworks = append(artworks, artworksForURL...)
	}
	if len(errs) > 0 {
		return nil, fmt.Errorf("fetching twitter encountered %d errors: %v", len(errs), errs)
	}
	return artworks, nil
}

func (t *Twitter) GetArtworkInfo(sourceURL string) (*types.Artwork, error) {
	tweetID := getTweetID(sourceURL)
	if tweetID == "" {
		return nil, ErrInvalidURL
	}
	fxTwitterApiURL := fmt.Sprintf("https://api.%s/_/status/%s", config.Cfg.Source.Twitter.FxTwitterDomain, tweetID)
	resp, err := reqApiResp(fxTwitterApiURL)
	if err != nil {
		return nil, err
	}
	return resp.ToArtwork()
}

func (t *Twitter) GetPictureInfo(sourceURL string, index uint) (*types.Picture, error) {
	artwork, err := t.GetArtworkInfo(sourceURL)
	if err != nil {
		return nil, err
	}
	if index >= uint(len(artwork.Pictures)) {
		return nil, ErrIndexOOB
	}
	return artwork.Pictures[index], nil
}

func (t *Twitter) GetSourceURLRegexp() *regexp.Regexp {
	return twitterSourceURLRegexp
}

func (t *Twitter) GetCommonSourceURL(twurl string) string {
	tweet := getTweetPath(twurl)
	if tweet == "" {
		return ""
	}
	commonUrl, err := url.JoinPath("https://x.com", tweet)
	if err != nil {
		common.Logger.Errorf("failed to join path: %v", err)
		return ""
	}
	return commonUrl
}

func (t *Twitter) GetFileName(artwork *types.Artwork, picture *types.Picture) string {
	original := picture.Original
	urlSplit := strings.Split(picture.Original, "?")
	if len(urlSplit) > 1 {
		original = strings.Join(urlSplit[:len(urlSplit)-1], "?")
	}
	tweetID := strings.Split(artwork.SourceURL, "/")[len(strings.Split(artwork.SourceURL, "/"))-1]
	return tweetID + "_" + strconv.Itoa(int(picture.Index)) + path.Ext(original)
}

func (t *Twitter) Config() *config.SourceCommonConfig {
	return &config.SourceCommonConfig{
		Enable:   config.Cfg.Source.Twitter.Enable,
		Intervel: config.Cfg.Source.Twitter.Intervel,
	}

}
