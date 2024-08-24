package storage

import (
	"ManyACG/common"
	"ManyACG/config"
	manyacgErrors "ManyACG/errors"
	. "ManyACG/logger"
	"ManyACG/sources"
	"ManyACG/storage/alist"
	"ManyACG/storage/local"
	"ManyACG/storage/webdav"
	"ManyACG/types"
	"context"
	"fmt"
	"path/filepath"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

var Storages = make(map[types.StorageType]Storage)

func InitStorage() {
	Logger.Info("Initializing storage")
	if config.Cfg.Storage.Local.Enable {
		Storages[types.StorageTypeLocal] = new(local.Local)
		Storages[types.StorageTypeLocal].Init()
	}
	if config.Cfg.Storage.Webdav.Enable {
		Storages[types.StorageTypeWebdav] = new(webdav.Webdav)
		Storages[types.StorageTypeWebdav].Init()
	}
	if config.Cfg.Storage.Alist.Enable {
		Storages[types.StorageTypeAlist] = new(alist.Alist)
		Storages[types.StorageTypeAlist].Init()
	}
}

// 将图片保存为所有尺寸
func SaveAll(ctx context.Context, artwork *types.Artwork, picture *types.Picture) (*types.StorageInfo, error) {
	Logger.Infof("saving picture %d of artwork %s", picture.Index, artwork.Title)
	originalBytes, err := common.DownloadWithCache(ctx, picture.Original, nil)
	if err != nil {
		return nil, err
	}
	filePath := config.Cfg.Storage.CacheDir + "/" + common.ReplaceFileNameInvalidChar(picture.Original)
	if err := common.MkFile(filePath, originalBytes); err != nil {
		return nil, err
	}
	defer func() {
		go common.PurgeFileAfter(filePath, time.Duration(config.Cfg.Storage.CacheTTL))
	}()
	originalStorageFileName, err := sources.GetFileName(artwork, picture)
	if err != nil {
		return nil, err
	}
	originalStoragePath := fmt.Sprintf("/%s/%s/%s", artwork.SourceType, common.ReplaceFileNameInvalidChar(artwork.Artist.Username), originalStorageFileName)
	originalStorage, ok := Storages[types.StorageType(config.Cfg.Storage.OriginalType)]
	if !ok {
		Logger.Fatalf("Unknown storage type: %s", config.Cfg.Storage.OriginalType)
		return nil, fmt.Errorf("%w: %s", manyacgErrors.ErrStorageUnkown, config.Cfg.Storage.OriginalType)
	}

	originalDetail, err := originalStorage.Save(ctx, filePath, originalStoragePath)
	if err != nil {
		return nil, err
	}

	var regularDetail *types.StorageDetail
	if config.Cfg.Storage.RegularType != "" {
		regularStorage, ok := Storages[types.StorageType(config.Cfg.Storage.RegularType)]
		if !ok {
			Logger.Fatalf("Unknown storage type: %s", config.Cfg.Storage.RegularType)
			return nil, fmt.Errorf("%w: %s", manyacgErrors.ErrStorageUnkown, config.Cfg.Storage.RegularType)
		}
		regularOutputPath := filePath[:len(filePath)-len(filepath.Ext(filePath))] + "_regular.webp"
		if err := common.CompressImageByFFmpeg(filePath, regularOutputPath, 2560, 75); err != nil {
			return nil, err
		}
		go common.PurgeFileAfter(regularOutputPath, time.Duration(config.Cfg.Storage.CacheTTL)*time.Second)

		if picture.ID == "" {
			picture.ID = primitive.NewObjectID().Hex()
		}
		regularStorageFileName := picture.ID + "_regular.webp"
		regularStoragePath := fmt.Sprintf("/regular/%s/%s/%s", artwork.SourceType, common.ReplaceFileNameInvalidChar(artwork.Artist.Username), regularStorageFileName)

		regularDetail, err = regularStorage.Save(ctx, regularOutputPath, regularStoragePath)
		if err != nil {
			return nil, err
		}
	}
	var thumbDetail *types.StorageDetail
	if config.Cfg.Storage.ThumbType != "" {
		thumbStorage, ok := Storages[types.StorageType(config.Cfg.Storage.ThumbType)]
		if !ok {
			Logger.Fatalf("Unknown storage type: %s", config.Cfg.Storage.ThumbType)
			return nil, fmt.Errorf("%w: %s", manyacgErrors.ErrStorageUnkown, config.Cfg.Storage.ThumbType)
		}
		thumbOutputPath := filePath[:len(filePath)-len(filepath.Ext(filePath))] + "_thumb.webp"
		if err := common.CompressImageByFFmpeg(filePath, thumbOutputPath, 500, 75); err != nil {
			return nil, err
		}
		go common.PurgeFileAfter(thumbOutputPath, time.Duration(config.Cfg.Storage.CacheTTL)*time.Second)
		if picture.ID == "" {
			picture.ID = primitive.NewObjectID().Hex()
		}
		thumbStorageFileName := picture.ID + "_thumb.webp"
		thumbStoragePath := fmt.Sprintf("/thumb/%s/%s/%s", artwork.SourceType, common.ReplaceFileNameInvalidChar(artwork.Artist.Username), thumbStorageFileName)

		thumbDetail, err = thumbStorage.Save(ctx, thumbOutputPath, thumbStoragePath)
		if err != nil {
			return nil, err
		}
	}
	return &types.StorageInfo{
		Original: originalDetail,
		Regular:  regularDetail,
		Thumb:    thumbDetail,
	}, nil
}

func Save(ctx context.Context, filePath string, storagePath string, storageType types.StorageType) (*types.StorageDetail, error) {
	if storage, ok := Storages[storageType]; ok {
		return storage.Save(ctx, filePath, storagePath)
	} else {
		return nil, fmt.Errorf("%w: %s", manyacgErrors.ErrStorageUnkown, storageType)
	}
}

func GetFile(ctx context.Context, detail *types.StorageDetail) ([]byte, error) {
	if storage, ok := Storages[detail.Type]; ok {
		return storage.GetFile(ctx, detail)
	} else {
		return nil, fmt.Errorf("%w: %s", manyacgErrors.ErrStorageUnkown, detail.Type)
	}
}

func Delete(ctx context.Context, info *types.StorageDetail) error {
	if storage, ok := Storages[info.Type]; ok {
		return storage.Delete(ctx, info)
	} else {
		return fmt.Errorf("%w: %s", manyacgErrors.ErrStorageUnkown, info.Type)
	}
}

func DeleteAll(ctx context.Context, info *types.StorageInfo) error {
	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	errChan := make(chan error)
	for _, detail := range []*types.StorageDetail{info.Original, info.Regular, info.Thumb} {
		if detail == nil {
			continue
		}
		wg.Add(1)
		go func(detail *types.StorageDetail) {
			defer wg.Done()
			if err := Delete(ctx, detail); err != nil {
				errChan <- err
				cancel()
			}
		}(detail)
	}
	go func() {
		wg.Wait()
		close(errChan)
	}()
	for err := range errChan {
		return err
	}
	return nil
}
