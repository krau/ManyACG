package repo

import (
	"context"

	"github.com/krau/ManyACG/internal/model/entity"
	"github.com/krau/ManyACG/pkg/objectuuid"
)

type Artwork interface {
	GetArtworkByID(ctx context.Context, id objectuuid.ObjectUUID) (*entity.Artwork, error)
	GetArtworkByURL(ctx context.Context, url string) (*entity.Artwork, error)
	CreateArtwork(ctx context.Context, artwork *entity.Artwork) (*objectuuid.ObjectUUID, error)
	UpdateArtworkByMap(ctx context.Context, id objectuuid.ObjectUUID, patch map[string]any) error
	UpdateArtworkTags(ctx context.Context, id objectuuid.ObjectUUID, tags []*entity.Tag) error
	UpdateArtworkPictures(ctx context.Context, id objectuuid.ObjectUUID, pictures []*entity.Picture) error
	ReorderArtworkPictures(ctx context.Context, id objectuuid.ObjectUUID) error
	DeleteArtworkByID(ctx context.Context, id objectuuid.ObjectUUID) error
}

type DeletedRecord interface {
	CheckDeletedByURL(ctx context.Context, url string) bool
}

type CachedArtwork interface {
	ResetPostingCachedArtworkStatus(ctx context.Context) error
}
