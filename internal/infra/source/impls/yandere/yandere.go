package yandere

import (
	"context"

	"github.com/imroc/req/v3"
	"github.com/krau/ManyACG/internal/infra/config"
	"github.com/krau/ManyACG/internal/infra/source"
	"github.com/krau/ManyACG/internal/shared"
)

type Yandere struct {
	cfg       config.SourceYandereConfig
	reqClient *req.Client
}

func init() {
	source.Register(shared.SourceTypeYandere, func() source.ArtworkSource {
		return &Yandere{
			cfg:       config.Get().Source.Yandere,
			reqClient: req.C().ImpersonateChrome().SetCommonRetryCount(2),
		}
	})
}

func (y *Yandere) FetchNewArtworks(ctx context.Context, limit int) ([]*source.FetchedArtwork, error) {
	return nil, nil
}

func (y *Yandere) GetArtworkInfo(ctx context.Context, sourceURL string) (*source.FetchedArtwork, error) {
	// postID := GetPostID(sourceURL)
	// if postID == "" {
	// 	return nil, ErrInvalidYanderePostURL
	// }
	// sourceURL = sourceURLPrefix + postID
	// common.Logger.Tracef("request artwork info: %s", sourceURL)
	// common.Logger.Tracef("getting yandere api: %s", apiBaseURL+postID)
	// resp, err := reqClient.R().Get(apiBaseURL + postID)
	// if err != nil {
	// 	return nil, err
	// }
	// if resp.IsErrorState() {
	// 	return nil, fmt.Errorf("%w: %s", ErrYandereAPIError, resp.Status)
	// }
	// var yandereResp YandereJsonResp
	// if err := json.Unmarshal(resp.Bytes(), &yandereResp); err != nil {
	// 	return nil, err
	// }
	// parentID := 0
	// if len(yandereResp) == 1 {
	// 	parentID = yandereResp[0].ParentID
	// }
	// if parentID == 0 {
	// 	return yandereResp.ToArtwork(), nil
	// }

	// parentURL := sourceURLPrefix + strconv.Itoa(parentID)
	// artwork, _ := service.GetArtworkByURL(context.TODO(), parentURL)
	// if artwork != nil {
	// 	return artwork, nil
	// }

	// apiURL := apiBaseURL + strconv.Itoa(parentID)
	// common.Logger.Tracef("getting yandere api: %s", apiURL)
	// resp, err = reqClient.R().Get(apiURL)
	// if err != nil {
	// 	return nil, err
	// }
	// if resp.IsErrorState() {
	// 	return nil, fmt.Errorf("%w: %s", ErrYandereAPIError, resp.Status)
	// }
	// var parentResp YandereJsonResp
	// if err := json.Unmarshal(resp.Bytes(), &parentResp); err != nil {
	// 	return nil, err
	// }

	// return parentResp.ToArtwork(), nil
	panic("not implemented")
}

func (y *Yandere) MatchesSourceURL(text string) (string, bool) {
	postID := GetPostID(text)
	if postID == "" {
		return "", false
	}
	return sourceURLPrefix + postID, true
}

// func (y *Yandere) GetFileName(artwork *types.Artwork, picture *types.Picture) string {
// 	idStr := GetPostID(artwork.SourceURL)
// 	if idStr == "" {
// 		idStr = picture.ID
// 	}
// 	if idStr == "" {
// 		idStr = common.MD5Hash(picture.Original)
// 	}
// 	return "yandere_" + idStr + "_" + strconv.Itoa(int(picture.Index)) + path.Ext(picture.Original)
// }

// func (y *Yandere) Config() *config.SourceCommonConfig {
// 	return &config.SourceCommonConfig{
// 		Enable:   config.Cfg.Source.Yandere.Enable,
// 		Intervel: -1,
// 	}
// }
