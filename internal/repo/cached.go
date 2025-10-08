package repo

import (
	"context"

	"github.com/krau/ManyACG/internal/model/entity"
	"github.com/krau/ManyACG/pkg/objectuuid"
)

type CachedArtwork interface {
	CreateCachedArtwork(ctx context.Context, cachedArt *entity.CachedArtwork) (*entity.CachedArtwork, error)
	ResetPostingCachedArtworkStatus(ctx context.Context) error
	DeleteCachedArtworkByID(ctx context.Context, id objectuuid.ObjectUUID) error
	GetCachedArtworkByURL(ctx context.Context, url string) (*entity.CachedArtwork, error)
	SaveCachedArtwork(ctx context.Context, artwork *entity.CachedArtwork) (*entity.CachedArtwork, error)
}
