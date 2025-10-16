package twitter

import (
	"context"
	"fmt"
	"net/url"
	"path"
	"strconv"
	"strings"

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

func Init() {
	cfg := runtimecfg.Get().Source.Twitter
	if cfg.Disable {
		return
	}
	source.Register(shared.SourceTypeTwitter, func() source.ArtworkSource {
		return &Twitter{
			cfg:       runtimecfg.Get().Source.Twitter,
			reqClient: req.C(),
		}
	})
}

func (t *Twitter) FetchNewArtworks(ctx context.Context, limit int) ([]*dto.FetchedArtwork, error) {
	return nil, nil
}

func (t *Twitter) GetArtworkInfo(ctx context.Context, sourceURL string) (*dto.FetchedArtwork, error) {
	tweetID := getTweetID(sourceURL)
	if tweetID == "" {
		return nil, ErrInvalidURL
	}
	fxTwitterApiURL := fmt.Sprintf("https://api.%s/_/status/%s", t.cfg.FxTwitterDomain, tweetID)
	resp, err := t.reqApiResp(ctx, fxTwitterApiURL)
	if err != nil {
		return nil, err
	}
	return resp.ToArtwork()
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

// PrettyFileName implements source.ArtworkSource.
func (t *Twitter) PrettyFileName(artwork shared.ArtworkLike, picture shared.PictureLike) string {
	original := picture.GetOriginal()
	urlSplit := strings.Split(original, "?")
	if len(urlSplit) > 1 {
		original = strings.Join(urlSplit[:len(urlSplit)-1], "?")
	}
	tweetID := strings.Split(artwork.GetSourceURL(), "/")[len(strings.Split(artwork.GetSourceURL(), "/"))-1]
	return "twitter_" + tweetID + "_" + strconv.Itoa(int(picture.GetIndex())) + path.Ext(original)
}
