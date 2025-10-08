package repo

import (
	"context"

	"github.com/krau/ManyACG/internal/model/entity"
)

type DeletedRecord interface {
	CheckDeletedByURL(ctx context.Context, url string) bool
	CreateDeletedRecord(ctx context.Context, record *entity.DeletedRecord) error
	// 删除不存在的记录不应返回错误.
	DeleteDeletedByURL(ctx context.Context, url string) error
	GetDeletedByURL(ctx context.Context, url string) (*entity.DeletedRecord, error)
}
