package storage

import (
	"errors"
	"path"
	"strings"

	"github.com/krau/ManyACG/config"
	"github.com/krau/ManyACG/types"
)

var (
	ErrNilStorageDetail = errors.New("storage detail is nil")
	ErrStorageNotFound  = errors.New("storage not found")
	ErrNoStorages       = errors.New("no storage found")
)

func applyRule(detail *types.StorageDetail) (*types.StorageDetail, error) {
	if detail == nil {
		return nil, ErrNilStorageDetail
	}

	currentType := detail.Type.String()
	currentPath := detail.Path

	if currentType == "" || currentPath == "" {
		return detail, nil
	}

	newValue := &types.StorageDetail{}
	for _, rule := range config.Cfg.Storage.Rules {
		if !(currentType == rule.StorageType && strings.HasPrefix(currentPath, rule.PathPrefix)) {
			continue
		}
		if rule.RewriteStorage == "" {
			continue
		}
		_, ok := Storages[types.StorageType(rule.RewriteStorage)]
		if !ok {
			return nil, ErrStorageNotFound
		}
		newValue.Type = types.StorageType(rule.RewriteStorage)
		newValue.Path = path.Join(rule.JoinPrefix, strings.TrimPrefix(currentPath, rule.TrimPrefix))
	}
	if newValue.Type == "" {
		return detail, nil
	}

	return newValue, nil
}
