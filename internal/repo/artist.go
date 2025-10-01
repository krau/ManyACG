package repo

import (
	"context"

	"github.com/krau/ManyACG/internal/model/entity"
	"github.com/krau/ManyACG/internal/shared"
	"github.com/krau/ManyACG/pkg/objectuuid"
)

type Artist interface {
	GetArtistByID(ctx context.Context, id objectuuid.ObjectUUID) (*entity.Artist, error)
	GetArtistByUID(ctx context.Context, uid string, sourceType shared.SourceType) (*entity.Artist, error)
	UpdateArtist(ctx context.Context, patch *entity.Artist) error
	CreateArtist(ctx context.Context, artist *entity.Artist) (*objectuuid.ObjectUUID, error)
}
