package service

import (
	"context"

	"github.com/krau/ManyACG/internal/model/entity"
	"github.com/krau/ManyACG/internal/shared"
	"gorm.io/datatypes"
)

func (s *Service) CreateApiKey(ctx context.Context, key string, quota int, permissions []shared.Permission, description string) (*entity.ApiKey, error) {
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
	apiKey := &entity.ApiKey{
		Key:         key,
		Quota:       quota,
		Used:        0,
		Permissions: datatypes.NewJSONSlice(permissions),
		Description: description,
	}
	_, err := s.repos.APIKey().CreateApiKey(ctx, apiKey)
	if err != nil {
		return nil, err
	}
	return s.repos.APIKey().GetApiKeyByKey(ctx, key)
}

// func GetApiKeyByKey(ctx context.Context, key string) (*entity.ApiKey, error) {
// 	// return dao.GetApiKeyByKey(ctx, key)
// 	return database.Default().GetApiKeyByKey(ctx, key)
// }

// func IncreaseApiKeyUsed(ctx context.Context, key string) error {
// 	// return dao.IncreaseApiKeyUsed(ctx, key)
// 	return database.Default().IncreaseApiKeyUsed(ctx, key)
// }

// func AddApiKeyQuota(ctx context.Context, key string, quota int) error {
// 	// return dao.AddApiKeyQuota(ctx, key, quota)
// 	return database.Default().AddApiKeyQuota(ctx, key, quota)
// }

// func DeleteApiKey(ctx context.Context, key string) error {
// 	// return dao.DeleteApiKey(ctx, key)
// 	return database.Default().DeleteApiKey(ctx, key)
// }
