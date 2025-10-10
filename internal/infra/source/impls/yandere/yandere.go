package yandere

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/imroc/req/v3"
	"github.com/krau/ManyACG/internal/infra/config/runtimecfg"
	"github.com/krau/ManyACG/internal/infra/source"
	"github.com/krau/ManyACG/internal/model/dto"
	"github.com/krau/ManyACG/internal/shared"
	"github.com/krau/ManyACG/pkg/strutil"
)

type Yandere struct {
	cfg       runtimecfg.SourceYandereConfig
	reqClient *req.Client
}

func init() {
	cfg := runtimecfg.Get().Source.Yandere
	if cfg.Disable {
		return
	}
	source.Register(shared.SourceTypeYandere, func() source.ArtworkSource {
		return &Yandere{
			cfg:       runtimecfg.Get().Source.Yandere,
			reqClient: req.C().ImpersonateChrome().SetCommonRetryCount(2),
		}
	})
}

func (y *Yandere) FetchNewArtworks(ctx context.Context, limit int) ([]*dto.FetchedArtwork, error) {
	return nil, nil
}

func (y *Yandere) GetArtworkInfo(ctx context.Context, sourceURL string) (*dto.FetchedArtwork, error) {
	postID := GetPostID(sourceURL)
	if postID == "" {
		return nil, ErrInvalidYanderePostURL
	}
	resp, err := y.reqClient.R().SetContext(ctx).Get(apiBaseURL + postID)
	if err != nil {
		return nil, err
	}
	if resp.IsErrorState() {
		return nil, fmt.Errorf("%w: %s", ErrYandereAPIError, resp.Status)
	}
	var yandereResp YandereJsonResp
	if err := json.Unmarshal(resp.Bytes(), &yandereResp); err != nil {
		return nil, err
	}
	parentID := 0
	if len(yandereResp) == 1 {
		parentID = yandereResp[0].ParentID
	}
	if parentID == 0 {
		return yandereResp.ToArtwork(), nil
	}

	// parentURL := sourceURLPrefix + strconv.Itoa(parentID)
	// artwork, _ := service.GetArtworkByURL(context.TODO(), parentURL)
	// if artwork != nil {
	// 	return artwork, nil
	// }

	apiURL := apiBaseURL + strconv.Itoa(parentID)
	resp, err = y.reqClient.R().SetContext(ctx).Get(apiURL)
	if err != nil {
		return nil, err
	}
	if resp.IsErrorState() {
		return nil, fmt.Errorf("%w: %s", ErrYandereAPIError, resp.Status)
	}
	var parentResp YandereJsonResp
	if err := json.Unmarshal(resp.Bytes(), &parentResp); err != nil {
		return nil, err
	}

	return parentResp.ToArtwork(), nil
}

func (y *Yandere) MatchesSourceURL(text string) (string, bool) {
	postID := GetPostID(text)
	if postID == "" {
		return "", false
	}
	return sourceURLPrefix + postID, true
}

// PrettyFileName implements source.ArtworkSource.
func (y *Yandere) PrettyFileName(artwork shared.ArtworkLike, picture shared.PictureLike) string {
	idStr := GetPostID(artwork.GetSourceURL())
	ext, _ := strutil.GetFileExtFromURL(picture.GetOriginal())
	if idStr == "" {
		return fmt.Sprintf("yandere_%s%s", strutil.MD5Hash(picture.GetOriginal()), ext)
	}
	return fmt.Sprintf("yandere_%s_%d%s", idStr, picture.GetIndex(), ext)
}
