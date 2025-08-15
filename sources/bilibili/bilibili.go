package bilibili

import (
	"errors"
	"fmt"
	"path"
	"regexp"

	"github.com/krau/ManyACG/common"
	"github.com/krau/ManyACG/config"
	sourceCommon "github.com/krau/ManyACG/sources/common"
	"github.com/krau/ManyACG/types"

	"github.com/imroc/req/v3"
)

type Bilibili struct{}

func init() {
	sourceCommon.RegisterSource(types.SourceTypeBilibili, new(Bilibili))
}

func (b *Bilibili) Init(_ types.Service) {
	reqClient = req.C().ImpersonateChrome()
	if config.Cfg.Source.Proxy != "" {
		reqClient.SetProxyURL(config.Cfg.Source.Proxy)
	}
	reqClient.SetCommonRetryCount(3)
}

func (b *Bilibili) FetchNewArtworksWithCh(artworkCh chan *types.Artwork, limit int) error {
	return nil
}

func (b *Bilibili) FetchNewArtworks(limit int) ([]*types.Artwork, error) {
	return nil, nil
}

func (b *Bilibili) GetArtworkInfo(sourceURL string) (*types.Artwork, error) {
	dynamicID := getDynamicID(sourceURL)
	if dynamicID == "" {
		return nil, ErrInvalidURL
	}
	common.Logger.Tracef("request artwork info: https://t.bilibili.com/%s", dynamicID)
	var err error
	var desktopResp *BilibiliDesktopDynamicApiResp
	desktopResp, err = reqDesktopDynamicApiResp(dynamicID)
	if err == nil {
		var artwork *types.Artwork
		artwork, err = desktopResp.ToArtwork()
		if errors.Is(err, ErrInvalidURL) {
			return nil, err
		}
		if err == nil {
			return artwork, nil
		}
	}
	var webResp *BilibiliWebDynamicApiResp
	webResp, err = reqWebDynamicApiResp(dynamicID)
	if err == nil {
		var artwork *types.Artwork
		artwork, err = webResp.ToArtwork()
		if err == nil {
			return artwork, nil
		}
	}
	return nil, err
}

func (b *Bilibili) GetPictureInfo(sourceURL string, index uint) (*types.Picture, error) {
	artwork, err := b.GetArtworkInfo(sourceURL)
	if err != nil {
		return nil, err
	}
	if index >= uint(len(artwork.Pictures)) {
		return nil, ErrIndexOOB
	}
	return artwork.Pictures[index], nil
}

func (b *Bilibili) GetSourceURLRegexp() *regexp.Regexp {
	return dynamicURLRegexp
}

func (b *Bilibili) GetCommonSourceURL(url string) string {
	dynamicID := getDynamicID(url)
	if dynamicID == "" {
		return ""
	}
	return "https://t.bilibili.com/" + dynamicID
}

func (b *Bilibili) GetFileName(artwork *types.Artwork, picture *types.Picture) string {
	dynamicID := getDynamicID(artwork.SourceURL)
	return fmt.Sprintf("%s_%d%s", dynamicID, picture.Index, path.Ext(picture.Original))
}

func (b *Bilibili) Config() *config.SourceCommonConfig {
	return &config.SourceCommonConfig{
		Enable:   config.Cfg.Source.Bilibili.Enable,
		Intervel: -1,
	}
}
