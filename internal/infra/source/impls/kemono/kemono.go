package kemono

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/imroc/req/v3"

	config "github.com/krau/ManyACG/internal/infra/config/runtimecfg"
	"github.com/krau/ManyACG/internal/infra/source"
	"github.com/krau/ManyACG/internal/model/dto"
	"github.com/krau/ManyACG/internal/shared"
	"github.com/krau/ManyACG/pkg/strutil"
)

type Kemono struct {
	cfg       config.SourceKemonoConfig
	reqClient *req.Client
}

func Init() {
	cfg := config.Get().Source.Kemono
	if cfg.Disable {
		return
	}
	source.Register(shared.SourceTypeKemono, func() source.ArtworkSource {
		return &Kemono{
			cfg: config.Get().Source.Kemono,
			reqClient: req.C().SetCommonHeader("Accept", "text/css").EnableAutoDecompress().
				SetCommonRetryCount(3).SetTLSHandshakeTimeout(20 * time.Second),
		}
	})
}

func (k *Kemono) FetchNewArtworks(ctx context.Context, limit int) ([]*dto.FetchedArtwork, error) {
	return nil, nil
}

func (k *Kemono) GetArtworkInfo(ctx context.Context, sourceURL string) (*dto.FetchedArtwork, error) {
	kemonoPostURL := kemonoSourceURLRegex.FindString(sourceURL)
	if kemonoPostURL == "" {
		return nil, ErrInvalidKemonoPostURL
	}
	parts := strings.Split(kemonoPostURL, "/")
	if len(parts) < 5 {
		return nil, ErrInvalidKemonoPostURL
	}
	service := parts[1]
	userID, err := strconv.Atoi(parts[3])
	if err != nil {
		return nil, fmt.Errorf("invalid user id: %w", err)
	}
	postID, err := strconv.Atoi(parts[5])
	if err != nil {
		return nil, fmt.Errorf("invalid post id: %w", err)
	}
	sourceURL = fmt.Sprintf("https://kemono.cr/%s/user/%d/post/%d", service, userID, postID)
	postPath := getPostPath(sourceURL)
	apiURL := apiBaseURL + postPath
	resp, err := k.reqClient.R().SetContext(ctx).Get(apiURL)
	if err != nil {
		return nil, err
	}
	var kemonoResp KemonoPostResp
	if err := json.Unmarshal(resp.Bytes(), &kemonoResp); err != nil {
		return nil, err
	}
	return k.convertToFetchedArtwork(ctx, &kemonoResp)
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

// PrettyFileName implements source.ArtworkSource.
func (k *Kemono) PrettyFileName(artwork shared.ArtworkLike, picture shared.PictureLike) string {
	ext, _ := strutil.GetFileExtFromURL(picture.GetOriginal())
	defaultName := fmt.Sprintf("kemono_%s%s", strutil.MD5Hash(picture.GetOriginal()), ext)
	postUrl := kemonoSourceURLRegex.FindString(artwork.GetSourceURL())
	if postUrl == "" {
		return defaultName
	}
	parts := strings.Split(postUrl, "/")
	if len(parts) < 5 {
		return defaultName
	}
	service := parts[1]
	userID := parts[3]
	postID := parts[5]
	return fmt.Sprintf("kemono_%s_%s_%s_%d%s", service, userID, postID, picture.GetIndex(), ext)
}
