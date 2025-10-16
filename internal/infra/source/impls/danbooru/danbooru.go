package danbooru

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/imroc/req/v3"
	"github.com/samber/oops"

	"github.com/krau/ManyACG/internal/infra/config/runtimecfg"
	"github.com/krau/ManyACG/internal/infra/source"
	"github.com/krau/ManyACG/internal/model/dto"
	"github.com/krau/ManyACG/internal/shared"
	"github.com/krau/ManyACG/pkg/strutil"
)

type Danbooru struct {
	cfg       runtimecfg.SourceDanbooruConfig
	reqClient *req.Client
}

func Init() {
	cfg := runtimecfg.Get().Source.Danbooru
	if cfg.Disable {
		return
	}
	source.Register(shared.SourceTypeDanbooru, func() source.ArtworkSource {
		return &Danbooru{
			cfg:       runtimecfg.Get().Source.Danbooru,
			reqClient: req.C().SetCommonRetryCount(3),
		}
	})
}

func (d *Danbooru) FetchNewArtworks(ctx context.Context, limit int) ([]*dto.FetchedArtwork, error) {
	return nil, nil
}

func (d *Danbooru) GetArtworkInfo(ctx context.Context, sourceURL string) (*dto.FetchedArtwork, error) {
	postID := GetPostID(sourceURL)
	if postID == "" {
		return nil, ErrInvalidDanbooruPostURL
	}
	sourceURL = "https://danbooru.donmai.us/posts/" + postID
	resp, err := d.reqClient.R().SetContext(ctx).Get(sourceURL + ".json")
	if err != nil {
		return nil, err
	}
	var danbooruResp DanbooruJsonResp
	if err := json.Unmarshal(resp.Bytes(), &danbooruResp); err != nil {
		return nil, err
	}
	if danbooruResp.Error != "" {
		return nil, oops.Errorf("danbooru api error: %s", danbooruResp.Message)
	}
	return danbooruResp.ToArtwork(), nil
}

func (d *Danbooru) MatchesSourceURL(text string) (string, bool) {
	postID := GetPostID(text)
	if postID == "" {
		return "", false
	}
	return "https://danbooru.donmai.us/posts/" + postID, true
}

// PrettyFileName implements source.ArtworkSource.
func (d *Danbooru) PrettyFileName(artwork shared.ArtworkLike, picture shared.PictureLike) string {
	ext, _ := strutil.GetFileExtFromURL(picture.GetOriginal())
	id := GetPostID(artwork.GetSourceURL())
	if id == "" {
		return fmt.Sprintf("danbooru_%s%s", strutil.MD5Hash(picture.GetOriginal()), ext)
	}
	return fmt.Sprintf("danbooru_%s_%d%s", id, picture.GetIndex(), ext)
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
