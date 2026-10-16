package common

import (
	"context"
	"io"

	"github.com/gofiber/fiber/v3"
	"github.com/krau/ManyACG/internal/model/entity"
	"github.com/krau/ManyACG/internal/model/query"
	"github.com/krau/ManyACG/internal/shared"
	"github.com/krau/ManyACG/pkg/osutil"
	"github.com/unvgo/ouid"
)

const (
	StateKeyQueryService = "query_service"
	StateKeyWriteService = "write_service"
	StateKeyKVStore      = "kv_store"
	StateKeyConfig       = "cfg"
	StateKeyLogger       = "logger"
	StateKeyTelegramBot  = "telegrambot" // DO NOT USE MustGetState to get this value!
)

type ArtworkDownloader interface {
	DownloadOriginFile(ctx context.Context, url string) (*osutil.File, error)
}

type QueryService interface {
	ArtworkDownloader
	GetArtistByID(ctx context.Context, id ouid.OUID) (*entity.Artist, error)
	GetArtworkByID(ctx context.Context, id ouid.OUID) (*entity.Artwork, error)
	QueryArtworks(ctx context.Context, que query.ArtworksDB) ([]*entity.Artwork, error)
	CountArtworks(ctx context.Context, que query.ArtworksDB) (int64, error)
	FindSimilarArtworks(ctx context.Context, que *query.ArtworkSimilar) ([]*entity.Artwork, error)
	SearchArtworks(ctx context.Context, que *query.ArtworkSearch) ([]*entity.Artwork, error)
	GetTagByNameWithAlias(ctx context.Context, name string) (*entity.Tag, error)
	RandomTags(ctx context.Context, limit int) ([]*entity.Tag, error)
	GetPictureByID(ctx context.Context, id ouid.OUID) (*entity.Picture, error)
	RandomPictures(ctx context.Context, limit int) ([]*entity.Picture, error)
	StorageGetFile(ctx context.Context, detail shared.StorageDetail) (*osutil.File, error)
	StorageStreamFile(ctx context.Context, detail shared.StorageDetail, w io.Writer) error
	PrettyFileName(artwork shared.ArtworkLike, picture shared.PictureLike) string
}

type WriteService interface {
	GetApiKeyByKey(ctx context.Context, key string) (*entity.ApiKey, error)
	IncreaseApiKeyUsed(ctx context.Context, key string) error
	FindSourceURL(text string) string
	GetOrFetchCachedArtwork(ctx context.Context, sourceURL string) (*entity.CachedArtwork, error)
	PrettyFileName(artwork shared.ArtworkLike, picture shared.PictureLike) string
	SavePicture(ctx context.Context, pic *entity.Picture) error
}

type TelegramBot interface {
	SendArtworkInfo(ctx context.Context, sourceUrl string, chatID int64, appendCaption string)
}

func GetState[T any](ctx fiber.Ctx, key string) (T, bool) {
	val, ok := ctx.App().State().Get(key)
	if !ok {
		var zero T
		return zero, false
	}
	vv, ok := val.(T)
	if !ok {
		var zero T
		return zero, false
	}
	return vv, true
}

func MustGetState[T any](ctx fiber.Ctx, key string) T {
	val := ctx.App().State().MustGet(key)
	v, ok := val.(T)
	if !ok {
		panic("state: dependency type assertion failed!")
	}
	return v
}
