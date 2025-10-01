package storage

import (
	"errors"
	"path"
	"strings"

	config "github.com/krau/ManyACG/internal/infra/config/runtimecfg"
	"github.com/krau/ManyACG/internal/shared"
)

var (
	ErrNilStorageDetail = errors.New("storage detail is nil")
	ErrStorageNotFound  = errors.New("storage not found")
	ErrNoStorages       = errors.New("no storage found")
)

func applyRule(detail *shared.StorageDetail) (*shared.StorageDetail, error) {
	if detail == nil {
		return nil, ErrNilStorageDetail
	}

	currentType := string(detail.Type)
	currentPath := detail.Path

	if currentType == "" || currentPath == "" {
		return detail, nil
	}

	newValue := &shared.StorageDetail{}
	for _, rule := range config.Get().Storage.Rules {
		if !(currentType == rule.StorageType && strings.HasPrefix(currentPath, rule.PathPrefix)) {
			continue
		}
		if rule.RewriteStorage == "" {
			continue
		}
		_, ok := storages[shared.StorageType(rule.RewriteStorage)]
		if !ok {
			return nil, ErrStorageNotFound
		}
		newValue.Type = shared.StorageType(rule.RewriteStorage)
		newValue.Path = path.Join(rule.JoinPrefix, strings.TrimPrefix(currentPath, rule.TrimPrefix))
		break
	}
	if newValue.Type == "" {
		return detail, nil
	}

	return newValue, nil
}
