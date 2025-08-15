package kemono

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/imroc/req/v3"
	"github.com/krau/ManyACG/common"
	"github.com/krau/ManyACG/config"
	scomon "github.com/krau/ManyACG/sources/common"

	"github.com/krau/ManyACG/types"
)

type Kemono struct{}

func init() {
	scomon.RegisterSource(types.SourceTypeKemono, new(Kemono))
}

func (k *Kemono) Init(_ types.Service) {
	reqClient = req.C().ImpersonateChrome()
	if config.Cfg.Source.Kemono.Session != "" {
		reqClient.SetCommonCookies(&http.Cookie{
			Name:  "session",
			Value: config.Cfg.Source.Kemono.Session,
		})
	}
	if config.Cfg.Source.Proxy != "" {
		reqClient.SetProxyURL(config.Cfg.Source.Proxy)
	}
	reqClient.SetCommonRetryCount(3).SetTLSHandshakeTimeout(20 * time.Second)
}

func (k *Kemono) FetchNewArtworksWithCh(artworkCh chan *types.Artwork, limit int) error {
	return nil
}

func (k *Kemono) FetchNewArtworks(limit int) ([]*types.Artwork, error) {
	return nil, nil
}

func (k *Kemono) GetArtworkInfo(sourceURL string) (*types.Artwork, error) {
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
	common.Logger.Tracef("request artwork info: %s", apiURL)
	resp, err := reqClient.R().Get(apiURL)
	if err != nil {
		return nil, err
	}
	var kemonoResp KemonoPostResp
	if err := json.Unmarshal(resp.Bytes(), &kemonoResp); err != nil {
		return nil, err
	}
	return kemonoResp.ToArtwork()
}

func (k *Kemono) GetPictureInfo(sourceURL string, index uint) (*types.Picture, error) {
	artwork, err := k.GetArtworkInfo(sourceURL)
	if err != nil {
		return nil, err
	}
	if index >= uint(len(artwork.Pictures)) {
		return nil, ErrIndexOOB
	}
	return artwork.Pictures[index], nil
}

func (k *Kemono) GetSourceURLRegexp() *regexp.Regexp {
	return kemonoSourceURLRegex
}

func (k *Kemono) GetCommonSourceURL(url string) string {
	kemonoPostURL := kemonoSourceURLRegex.FindString(url)
	if kemonoPostURL == "" {
		return ""
	}
	parts := strings.Split(kemonoPostURL, "/")
	if len(parts) < 5 {
		common.Logger.Fatalf("invalid kemono post url: %s", kemonoPostURL)
		return ""
	}
	service := parts[1]
	userID := parts[3]
	postID := parts[5]
	return fmt.Sprintf("https://kemono.cr/%s/user/%s/post/%s", service, userID, postID)
}

func (k *Kemono) GetFileName(artwork *types.Artwork, picture *types.Picture) string {
	return artwork.Title + "_" + strconv.Itoa(int(picture.Index)) + "_" + path.Base(picture.Original)
}

func (k *Kemono) Config() *config.SourceCommonConfig {
	return &config.SourceCommonConfig{
		Enable:   config.Cfg.Source.Kemono.Enable,
		Intervel: -1,
	}
}
