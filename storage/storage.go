package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"sync"
	"time"

	"github.com/krau/ManyACG/common"
	"github.com/krau/ManyACG/config"
	"github.com/krau/ManyACG/errs"

	"github.com/krau/ManyACG/sources"
	"github.com/krau/ManyACG/storage/alist"
	"github.com/krau/ManyACG/storage/local"
	"github.com/krau/ManyACG/storage/webdav"
	"github.com/krau/ManyACG/types"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

var Storages = make(map[types.StorageType]Storage)

func InitStorage() {
	common.Logger.Info("Initializing storage")
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
	if len(Storages) == 0 {
		return &types.StorageInfo{
			Original: nil,
			Regular:  nil,
			Thumb:    nil,
		}, nil
	}
	common.Logger.Infof("saving picture %d of artwork %s", picture.Index, artwork.Title)
	originalBytes, err := common.DownloadWithCache(ctx, picture.Original, nil)
	if err != nil {
		return nil, err
	}
	filePath := filepath.Join(config.Cfg.Storage.CacheDir, common.MD5Hash(picture.Original))
	if err := common.MkFile(filePath, originalBytes); err != nil {
		return nil, err
	}
	defer func() {
		go common.RmFileAfter(filePath, time.Duration(config.Cfg.Storage.CacheTTL)*time.Second)
	}()
	originalStorageFileName, err := sources.GetFileName(artwork, picture)
	if err != nil {
		return nil, err
	}
	originalStoragePath := fmt.Sprintf("/%s/%s/%s", artwork.SourceType, artwork.Artist.UID, originalStorageFileName)
	originalStorage, ok := Storages[types.StorageType(config.Cfg.Storage.OriginalType)]
	if !ok {
		common.Logger.Fatalf("Unknown storage type: %s", config.Cfg.Storage.OriginalType)
		return nil, fmt.Errorf("%w: %s", errs.ErrStorageUnkown, config.Cfg.Storage.OriginalType)
	}

	originalDetail, err := originalStorage.Save(ctx, filePath, originalStoragePath)
	if err != nil {
		return nil, err
	}

	var regularDetail *types.StorageDetail
	if config.Cfg.Storage.RegularType != "" {
		regularStorage, ok := Storages[types.StorageType(config.Cfg.Storage.RegularType)]
		if !ok {
			common.Logger.Fatalf("Unknown storage type: %s", config.Cfg.Storage.RegularType)
			return nil, fmt.Errorf("%w: %s", errs.ErrStorageUnkown, config.Cfg.Storage.RegularType)
		}
		regularOutputPath := fmt.Sprintf("%s_regular.%s", filePath[:len(filePath)-len(filepath.Ext(filePath))], config.Cfg.Storage.RegularFormat)
		if err := common.CompressImageByFFmpeg(filePath, regularOutputPath, types.RegularPhotoSideLength); err != nil {
			return nil, err
		}
		defer func() {
			go common.RmFileAfter(regularOutputPath, time.Duration(config.Cfg.Storage.CacheTTL)*time.Second)
		}()

		if picture.ID == "" {
			picture.ID = primitive.NewObjectID().Hex()
		}
		regularStorageFileName := picture.ID + "_regular." + config.Cfg.Storage.RegularFormat
		regularStoragePath := fmt.Sprintf("/regular/%s/%s/%s", artwork.SourceType, artwork.Artist.UID, regularStorageFileName)

		regularDetail, err = regularStorage.Save(ctx, regularOutputPath, regularStoragePath)
		if err != nil {
			return nil, err
		}
	}
	var thumbDetail *types.StorageDetail
	if config.Cfg.Storage.ThumbType != "" {
		thumbStorage, ok := Storages[types.StorageType(config.Cfg.Storage.ThumbType)]
		if !ok {
			common.Logger.Fatalf("Unknown storage type: %s", config.Cfg.Storage.ThumbType)
			return nil, fmt.Errorf("%w: %s", errs.ErrStorageUnkown, config.Cfg.Storage.ThumbType)
		}
		// thumbOutputPath := filePath[:len(filePath)-len(filepath.Ext(filePath))] + "_thumb.webp"
		thumbOutputPath := fmt.Sprintf("%s_thumb.webp", filePath[:len(filePath)-len(filepath.Ext(filePath))])
		if err := common.CompressImageByFFmpeg(filePath, thumbOutputPath, types.ThumbPhotoSideLength); err != nil {
			return nil, err
		}

		defer func() {
			go common.RmFileAfter(thumbOutputPath, time.Duration(config.Cfg.Storage.CacheTTL)*time.Second)
		}()

		if picture.ID == "" {
			picture.ID = primitive.NewObjectID().Hex()
		}
		thumbStorageFileName := picture.ID + "_thumb." + config.Cfg.Storage.ThumbFormat
		thumbStoragePath := fmt.Sprintf("/thumb/%s/%s/%s", artwork.SourceType, artwork.Artist.UID, thumbStorageFileName)

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
		return nil, fmt.Errorf("%w: %s", errs.ErrStorageUnkown, storageType)
	}
}

var storageLocks sync.Map

func GetFile(ctx context.Context, detail *types.StorageDetail) ([]byte, error) {
	if detail.Type != types.StorageTypeLocal {
		lock, _ := storageLocks.LoadOrStore(detail.String(), &sync.Mutex{})
		lock.(*sync.Mutex).Lock()
		defer func() {
			lock.(*sync.Mutex).Unlock()
			storageLocks.Delete(detail)
		}()
	}
	if storage, ok := Storages[detail.Type]; ok {
		file, err := storage.GetFile(ctx, detail)
		if err != nil {
			return nil, err
		}
		return file, nil
	} else {
		return nil, fmt.Errorf("%w: %s", errs.ErrStorageUnkown, detail.Type)
	}
}

func GetFileStream(ctx context.Context, detail *types.StorageDetail) (io.ReadCloser, error) {
	if detail == nil {
		return nil, errors.New("storage detail is nil")
	}
	if detail.Type != types.StorageTypeLocal {
		lock, _ := storageLocks.LoadOrStore(detail.String(), &sync.Mutex{})
		lock.(*sync.Mutex).Lock()
		defer func() {
			lock.(*sync.Mutex).Unlock()
			storageLocks.Delete(detail)
		}()
	}
	if storage, ok := Storages[detail.Type]; ok {
		file, err := storage.GetFileStream(ctx, detail)
		if err != nil {
			return nil, err
		}
		return file, nil
	} else {
		return nil, fmt.Errorf("%w: %s", errs.ErrStorageUnkown, detail.Type)
	}
}

func Delete(ctx context.Context, info *types.StorageDetail) error {
	if storage, ok := Storages[info.Type]; ok {
		return storage.Delete(ctx, info)
	} else {
		return fmt.Errorf("%w: %s", errs.ErrStorageUnkown, info.Type)
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
