package repo

import (
	"context"

	"github.com/krau/ManyACG/internal/model/entity"
	"github.com/krau/ManyACG/internal/model/query"
	"github.com/krau/ManyACG/pkg/objectuuid"
)

type Artwork interface {
	GetArtworkByID(ctx context.Context, id objectuuid.ObjectUUID) (*entity.Artwork, error)
	GetArtworkByURL(ctx context.Context, url string) (*entity.Artwork, error)
	CreateArtwork(ctx context.Context, artwork *entity.Artwork) (*objectuuid.ObjectUUID, error)
	UpdateArtworkByMap(ctx context.Context, id objectuuid.ObjectUUID, patch map[string]any) error
	UpdateArtworkTags(ctx context.Context, id objectuuid.ObjectUUID, tags []*entity.Tag) error
	UpdateArtworkPictures(ctx context.Context, id objectuuid.ObjectUUID, pictures []*entity.Picture) error
	ReorderArtworkPicturesByID(ctx context.Context, id objectuuid.ObjectUUID) error
	DeleteArtworkByID(ctx context.Context, id objectuuid.ObjectUUID) error
	QueryArtworks(ctx context.Context, que query.ArtworksDB) ([]*entity.Artwork, error)
	GetArtworksByIDs(ctx context.Context, ids []objectuuid.ObjectUUID) ([]*entity.Artwork, error)
}

type DeletedRecord interface {
	CheckDeletedByURL(ctx context.Context, url string) bool
	CreateDeletedRecord(ctx context.Context, record *entity.DeletedRecord) error
	// 删除不存在的记录不应返回错误.
	DeleteDeletedByURL(ctx context.Context, url string) error
	GetDeletedByURL(ctx context.Context, url string) (*entity.DeletedRecord, error)
}

type CachedArtwork interface {
	CreateCachedArtwork(ctx context.Context, cachedArt *entity.CachedArtwork) (*entity.CachedArtwork, error)
	ResetPostingCachedArtworkStatus(ctx context.Context) error
	DeleteCachedArtworkByID(ctx context.Context, id objectuuid.ObjectUUID) error
	GetCachedArtworkByID(ctx context.Context, id objectuuid.ObjectUUID) (*entity.CachedArtwork, error)
	GetCachedArtworkByURL(ctx context.Context, url string) (*entity.CachedArtwork, error)
	SaveCachedArtwork(ctx context.Context, artwork *entity.CachedArtwork) (*entity.CachedArtwork, error)
}
