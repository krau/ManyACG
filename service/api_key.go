package service

import (
	"context"

	"github.com/krau/ManyACG/internal/infra/database"
	"github.com/krau/ManyACG/internal/infra/database/model"
	"github.com/krau/ManyACG/types"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func CreateApiKey(ctx context.Context, key string, quota int, permissions []types.ApiKeyPermission, description string) (*types.ApiKeyModel, error) {
	// apiKey := &types.ApiKeyModel{
	// 	Key:         key,
	// 	Quota:       quota,
	// 	Used:        0,
	// 	Permissions: permissions,
	// 	Description: description,
	// }
	// _, err := dao.CreateApiKey(ctx, apiKey)
	// if err != nil {
	// 	return nil, err
	// }
	// return dao.GetApiKeyByKey(ctx, key)
	apiKey := &model.ApiKey{
		Key:   key,
		Quota: quota,
		Used:  0,
		Permissions: func() []string {
			var perms []string
			for _, p := range permissions {
				perms = append(perms, string(p))
			}
			return perms
		}(),
		Description: description,
	}
	_, err := database.Default().CreateApiKey(ctx, apiKey)
	if err != nil {
		return nil, err
	}
	id, err := database.Default().GetApiKeyByKey(ctx, key)
	if err != nil {
		return nil, err
	}
	return &types.ApiKeyModel{
		ID:    primitive.ObjectID(id.ID.ToObjectID()),
		Key:   id.Key,
		Quota: id.Quota,
		Used:  id.Used,
		Permissions: func() []types.ApiKeyPermission {
			var perms []types.ApiKeyPermission
			for _, p := range id.Permissions {
				perms = append(perms, types.ApiKeyPermission(p))
			}
			return perms
		}(),
		Description: id.Description,
	}, nil
}

func GetApiKeyByKey(ctx context.Context, key string) (*types.ApiKeyModel, error) {
	// return dao.GetApiKeyByKey(ctx, key)
	apiKey, err := database.Default().GetApiKeyByKey(ctx, key)
	if err != nil {
		return nil, err
	}
	return &types.ApiKeyModel{
		ID:    primitive.ObjectID(apiKey.ID.ToObjectID()),
		Key:   apiKey.Key,
		Quota: apiKey.Quota,
		Used:  apiKey.Used,
		Permissions: func() []types.ApiKeyPermission {
			var perms []types.ApiKeyPermission
			for _, p := range apiKey.Permissions {
				perms = append(perms, types.ApiKeyPermission(p))
			}
			return perms
		}(),
		Description: apiKey.Description,
	}, nil
}

func IncreaseApiKeyUsed(ctx context.Context, key string) error {
	// return dao.IncreaseApiKeyUsed(ctx, key)
	return database.Default().IncreaseApiKeyUsed(ctx, key)
}

func AddApiKeyQuota(ctx context.Context, key string, quota int) error {
	// return dao.AddApiKeyQuota(ctx, key, quota)
	return database.Default().AddApiKeyQuota(ctx, key, quota)
}

func DeleteApiKey(ctx context.Context, key string) error {
	// return dao.DeleteApiKey(ctx, key)
	return database.Default().DeleteApiKey(ctx, key)
}
