package danbooru

import (
	"context"

	"github.com/imroc/req/v3"

	"github.com/krau/ManyACG/internal/infra/config"
	"github.com/krau/ManyACG/internal/infra/source"
	"github.com/krau/ManyACG/internal/shared"
)

type Danbooru struct {
	cfg       config.SourceDanbooruConfig
	reqClient *req.Client
}

func init() {
	source.Register(shared.SourceTypeDanbooru, func() source.ArtworkSource {
		return &Danbooru{
			cfg:       config.Get().Source.Danbooru,
			reqClient: req.C().SetCommonRetryCount(3),
		}
	})
}

func (d *Danbooru) FetchNewArtworks(ctx context.Context, limit int) ([]*source.FetchedArtwork, error) {
	return nil, nil
}

func (d *Danbooru) GetArtworkInfo(ctx context.Context, sourceURL string) (*source.FetchedArtwork, error) {
	// postID := GetPostID(sourceURL)
	// if postID == "" {
	// 	return nil, ErrInvalidDanbooruPostURL
	// }
	// sourceURL = "https://danbooru.donmai.us/posts/" + postID
	// resp, err := reqClient.R().Get(sourceURL + ".json")
	// if err != nil {
	// 	return nil, err
	// }
	// var danbooruResp DanbooruJsonResp
	// if err := json.Unmarshal(resp.Bytes(), &danbooruResp); err != nil {
	// 	return nil, err
	// }
	// if danbooruResp.Error != "" {
	// 	return nil, errors.New(danbooruResp.Message)
	// }
	// return danbooruResp.ToArtwork(), nil
	panic("not implemented")
}

func (d *Danbooru) MatchesSourceURL(text string) (string, bool) {
	postID := GetPostID(text)
	if postID == "" {
		return "", false
	}
	return "https://danbooru.donmai.us/posts/" + postID, true
}

// func (d *Danbooru) GetFileName(artwork *types.Artwork, picture *types.Picture) string {
// 	idStr := GetPostID(artwork.SourceURL)
// 	return fmt.Sprintf("%s_%d%s", idStr, picture.Index, path.Ext(picture.Original))
// }

// func (d *Danbooru) Config() *config.SourceCommonConfig {
// 	return &config.SourceCommonConfig{
// 		Enable:   config.Get().Source.Danbooru.Enable,
// 		Intervel: -1,
// 	}
// }
