package repo

import (
	"context"

	"github.com/krau/ManyACG/internal/model/entity"
	"github.com/unvgo/ouid"
)

type APIKey interface {
	CreateApiKey(ctx context.Context, apikey *entity.ApiKey) (*ouid.OUID, error)
	GetApiKeyByKey(ctx context.Context, key string) (*entity.ApiKey, error)
	IncreaseApiKeyUsed(ctx context.Context, key string) error
	AddApiKeyQuota(ctx context.Context, key string, quota int) error
	DeleteApiKey(ctx context.Context, key string) error
}
