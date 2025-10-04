package service

import (
	"context"

	"github.com/krau/ManyACG/internal/infra/storage"
	"github.com/krau/ManyACG/internal/shared"
	"github.com/samber/oops"
)

func (s *Service) Storage(storageType shared.StorageType) storage.Storage {
	return s.storages[storageType]
}

func (s *Service) StorageDelete(ctx context.Context, detail shared.StorageDetail) error {
	if stor, ok := s.storages[detail.Type]; ok {
		return stor.Delete(ctx, detail)
	}
	return oops.Errorf("storage type %s not found", detail.Type)
}

func (s *Service) StorageDeleteByInfo(ctx context.Context, info shared.StorageInfo) error {
	if info.Original != nil {
		if err := s.StorageDelete(ctx, *info.Original); err != nil {
			return err
		}
	}
	if info.Regular != nil {
		if err := s.StorageDelete(ctx, *info.Regular); err != nil {
			return err
		}
	}
	if info.Thumb != nil {
		if err := s.StorageDelete(ctx, *info.Thumb); err != nil {
			return err
		}
	}
	return nil
}
