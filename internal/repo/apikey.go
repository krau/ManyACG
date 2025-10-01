package repo

import (
	"context"

	"github.com/krau/ManyACG/internal/model/entity"
	"github.com/krau/ManyACG/pkg/objectuuid"
)

type APIKey interface {
	CreateApiKey(ctx context.Context, apikey *entity.ApiKey) (*objectuuid.ObjectUUID, error)
	GetApiKeyByKey(ctx context.Context, key string) (*entity.ApiKey, error)
}
