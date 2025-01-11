package service

import (
	"context"

	"github.com/krau/ManyACG/dao"
	"github.com/krau/ManyACG/types"
)

func CreateApiKey(ctx context.Context, key string, quota int, permissions []types.ApiKeyPermission) (*types.ApiKeyModel, error) {
	apiKey := &types.ApiKeyModel{
		Key:         key,
		Quota:       quota,
		Used:        0,
		Permissions: permissions,
	}
	_, err := dao.CreateApiKey(ctx, apiKey)
	if err != nil {
		return nil, err
	}
	return dao.GetApiKeyByKey(ctx, key)
}

func GetApiKeyByKey(ctx context.Context, key string) (*types.ApiKeyModel, error) {
	return dao.GetApiKeyByKey(ctx, key)
}

func IncreaseApiKeyUsed(ctx context.Context, key string) error {
	return dao.IncreaseApiKeyUsed(ctx, key)
}

func AddApiKeyQuota(ctx context.Context, key string, quota int) error {
	return dao.AddApiKeyQuota(ctx, key, quota)
}

func DeleteApiKey(ctx context.Context, key string) error {
	return dao.DeleteApiKey(ctx, key)
}
