package repo

import (
	"context"

	"github.com/krau/ManyACG/internal/model/entity"
	"github.com/unvgo/ouid"
)

type Tag interface {
	GetTagByName(ctx context.Context, name string) (*entity.Tag, error)
	GetTagByNameWithAlias(ctx context.Context, name string) (*entity.Tag, error)
	GetAliasTagByName(ctx context.Context, name string) (*entity.TagAlias, error)
	GetTagByID(ctx context.Context, id ouid.OUID) (*entity.Tag, error)
	CreateTag(ctx context.Context, tag *entity.Tag) (*entity.Tag, error)
	RandomTags(ctx context.Context, limit int) ([]*entity.Tag, error)
	// MigrateTagAlias 将 aliasTagID 迁移到 targetTagID，并返回受影响的 artwork ID 列表
	MigrateTagAlias(ctx context.Context, aliasTagID, targetTagID ouid.OUID) ([]ouid.OUID, error)
	UpdateTagAlias(ctx context.Context, id ouid.OUID, alias []*entity.TagAlias) error
}
