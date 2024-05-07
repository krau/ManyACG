package twitter

import (
	"ManyACG-Bot/config"
	"ManyACG-Bot/types"
)

type Twitter struct{}

func (t *Twitter) Init() {
}

func (t *Twitter) FetchNewArtworksWithCh(artworkCh chan *types.Artwork, limit int) error {
	return nil
}

func (t *Twitter) FetchNewArtworks(limit int) ([]*types.Artwork, error) {
	return nil, nil
}

func (t *Twitter) GetArtworkInfo(sourceURL string) (*types.Artwork, error) {
	tweetPath := GetTweetPath(sourceURL)
	if tweetPath == "" {
		return nil, ErrInvalidURL
	}
	fxTwitterApiURL := "https://api." + config.Cfg.Source.Twitter.FxTwitterDomain + "/" + tweetPath
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

func (t *Twitter) Config() *config.SourceCommonConfig {
	return &config.SourceCommonConfig{
		Enable:   config.Cfg.Source.Twitter.Enable,
		Intervel: -1, // Twitter 暂无法实现主动抓取的功能
	}

}
