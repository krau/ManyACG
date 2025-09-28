package kemono

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/imroc/req/v3"

	"github.com/krau/ManyACG/internal/infra/config"
	"github.com/krau/ManyACG/internal/infra/source"
	"github.com/krau/ManyACG/internal/shared"
)

type Kemono struct {
	cfg       config.SourceKemonoConfig
	reqClient *req.Client
}

func init() {
	source.Register(shared.SourceTypeKemono, func() source.ArtworkSource {
		return &Kemono{
			cfg: config.Get().Source.Kemono,
			reqClient: req.C().SetCommonHeader("Accept", "text/css").EnableAutoDecompress().
				SetCommonRetryCount(3).SetTLSHandshakeTimeout(20 * time.Second),
		}
	})
}

func (k *Kemono) FetchNewArtworks(ctx context.Context, limit int) ([]*source.FetchedArtwork, error) {
	return nil, nil
}

func (k *Kemono) GetArtworkInfo(ctx context.Context, sourceURL string) (*source.FetchedArtwork, error) {
	// kemonoPostURL := kemonoSourceURLRegex.FindString(sourceURL)
	// if kemonoPostURL == "" {
	// 	return nil, ErrInvalidKemonoPostURL
	// }
	// parts := strings.Split(kemonoPostURL, "/")
	// if len(parts) < 5 {
	// 	return nil, ErrInvalidKemonoPostURL
	// }
	// service := parts[1]
	// userID, err := strconv.Atoi(parts[3])
	// if err != nil {
	// 	return nil, fmt.Errorf("invalid user id: %w", err)
	// }
	// postID, err := strconv.Atoi(parts[5])
	// if err != nil {
	// 	return nil, fmt.Errorf("invalid post id: %w", err)
	// }
	// sourceURL = fmt.Sprintf("https://kemono.cr/%s/user/%d/post/%d", service, userID, postID)
	// postPath := getPostPath(sourceURL)
	// apiURL := apiBaseURL + postPath
	// common.Logger.Tracef("request artwork info: %s", apiURL)
	// resp, err := reqClient.R().Get(apiURL)
	// if err != nil {
	// 	return nil, err
	// }
	// var kemonoResp KemonoPostResp
	// if err := json.Unmarshal(resp.Bytes(), &kemonoResp); err != nil {
	// 	return nil, err
	// }
	// return kemonoResp.ToArtwork()
	panic("not implemented")
}

func (k *Kemono) MatchesSourceURL(text string) (string, bool) {
	kemonoPostURL := kemonoSourceURLRegex.FindString(text)
	if kemonoPostURL == "" {
		return "", false
	}
	parts := strings.Split(kemonoPostURL, "/")
	if len(parts) < 5 {
		return "", false
	}
	service := parts[1]
	userID := parts[3]
	postID := parts[5]
	return fmt.Sprintf("https://kemono.cr/%s/user/%s/post/%s", service, userID, postID), true
}

// func (k *Kemono) GetFileName(artwork *types.Artwork, picture *types.Picture) string {
// 	return artwork.Title + "_" + strconv.Itoa(int(picture.Index)) + "_" + path.Base(picture.Original)
// }

// func (k *Kemono) Config() *config.SourceCommonConfig {
// 	return &config.SourceCommonConfig{
// 		Enable:   config.Cfg.Source.Kemono.Enable,
// 		Intervel: -1,
// 	}
// }
