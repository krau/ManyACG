package repo

import (
	"context"

	"github.com/krau/ManyACG/internal/model/entity"
	"github.com/krau/ManyACG/pkg/objectuuid"
)

type Tag interface {
	GetTagByName(ctx context.Context, name string) (*entity.Tag, error)
	GetAliasTagByName(ctx context.Context, name string) (*entity.TagAlias, error)
	GetTagByID(ctx context.Context, id objectuuid.ObjectUUID) (*entity.Tag, error)
	CreateTag(ctx context.Context, tag *entity.Tag) (*objectuuid.ObjectUUID, error)
	MigrateTagAlias(ctx context.Context, aliasTagID, targetTagID objectuuid.ObjectUUID) error
	UpdateTagAlias(ctx context.Context, id objectuuid.ObjectUUID, alias []*entity.TagAlias) error
}
