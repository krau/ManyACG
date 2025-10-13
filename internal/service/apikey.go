package service

import (
	"context"

	"github.com/krau/ManyACG/internal/model/entity"
	"github.com/krau/ManyACG/internal/shared"
	"gorm.io/datatypes"
)

func (s *Service) CreateApiKey(ctx context.Context, key string, quota int, permissions []shared.Permission, description string) (*entity.ApiKey, error) {
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

func (s *Service) GetApiKeyByKey(ctx context.Context, key string) (*entity.ApiKey, error) {
	return s.repos.APIKey().GetApiKeyByKey(ctx, key)
}

func (s *Service) IncreaseApiKeyUsed(ctx context.Context, key string) error {
	return s.repos.APIKey().IncreaseApiKeyUsed(ctx, key)
}

func (s *Service) AddApiKeyQuota(ctx context.Context, key string, quota int) error {
	return s.repos.APIKey().AddApiKeyQuota(ctx, key, quota)
}

func (s *Service) DeleteApiKey(ctx context.Context, key string) error {
	return s.repos.APIKey().DeleteApiKey(ctx, key)
}
