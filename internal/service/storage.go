package service

import (
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/gabriel-vasile/mimetype"
	"github.com/krau/ManyACG/internal/infra/cache"
	"github.com/krau/ManyACG/internal/pkg/imgtool"
	"github.com/krau/ManyACG/internal/shared"
	"github.com/krau/ManyACG/pkg/osutil"
	"github.com/samber/oops"
)

func (s *Service) storageCachePath(detail shared.StorageDetail) string {
	cachePath := filepath.Join(s.storCfg.CacheDir, "storage", detail.Hash())
	ext := filepath.Ext(detail.Path)
	if ext != "" {
		cachePath += ext
	}
	return cachePath
}

func (s *Service) storageApplyRule(ctx context.Context, detail shared.StorageDetail) (shared.StorageDetail, error) {
	currentType := detail.Type.String()
	currentPath := detail.Path

	if currentType == "" || currentPath == "" {
		return detail, nil
	}
	cacheKey := fmt.Sprintf("storage:apply_rule:%s", detail.Hash())
	if cached, err := cache.Get[shared.StorageDetail](ctx, cacheKey); err == nil {
		return cached, nil
	}

	newValue := shared.StorageDetail{}
	for _, rule := range s.storCfg.Rules {
		if !(currentType == rule.MatchType && strings.HasPrefix(currentPath, rule.MatchPrefix)) {
			continue
		}
		if rule.RewriteStorage == "" {
			continue
		}
		newType, err := shared.ParseStorageType(rule.RewriteStorage)
		if err != nil {
			return shared.ZeroStorageDetail, oops.Wrapf(err, "parse storage type %s failed", rule.RewriteStorage)
		}
		_, ok := s.storages[newType]
		if !ok {
			return shared.ZeroStorageDetail, oops.Errorf("storage type %s not found", rule.RewriteStorage)
		}
		newValue.Type = newType
		newValue.Path = path.Join(rule.JoinPrefix, strings.TrimPrefix(currentPath, rule.TrimPrefix))
		break
	}
	if newValue.Type == "" {
		// no rule applied
		cache.Set(ctx, cacheKey, detail)
		return detail, nil
	}
	newValue.Mime = detail.Mime
	cache.Set(ctx, cacheKey, newValue)
	return newValue, nil

}

func (s *Service) StorageGetFile(ctx context.Context, detail shared.StorageDetail) (*osutil.File, error) {
	ruledDetail, err := s.storageApplyRule(ctx, detail)
	if err != nil {
		return nil, oops.Wrapf(err, "apply storage rule failed")
	}
	if ruledDetail.Type != "" && ruledDetail.Path != "" {
		detail = ruledDetail
	}

	cachePath := s.storageCachePath(detail)
	ext := filepath.Ext(detail.Path)
	if ext != "" {
		cachePath += ext
	}
	if stor, ok := s.storages[detail.Type]; ok {
		// 先检查缓存
		if cacheFile, err := osutil.OpenCache(cachePath); err == nil {
			return cacheFile, nil
		}
		// 从存储获取
		rc, err := stor.GetFile(ctx, detail)
		if err != nil {
			return nil, oops.Wrapf(err, "get file from storage %s failed", detail.Type)
		}
		defer rc.Close()
		// 读取到临时文件, 避免频繁从远程存储获取文件
		cacheFile, err := osutil.CreateCache(cachePath)
		if err != nil {
			return nil, oops.Wrapf(err, "create cache file failed")
		}
		if _, err := io.Copy(cacheFile, rc); err != nil {
			osutil.RemoveNow(cacheFile.Name())
			return nil, oops.Wrapf(err, "write cache file failed")
		}
		cacheFile.Seek(0, io.SeekStart)
		return cacheFile, nil
	}
	return nil, oops.Errorf("storage type %s not found", detail.Type)
}

func (s *Service) StorageStreamFile(ctx context.Context, detail shared.StorageDetail, w io.Writer) error {
	ruledDetail, err := s.storageApplyRule(ctx, detail)
	if err != nil {
		return oops.Wrapf(err, "apply storage rule failed")
	}
	if ruledDetail.Type != "" && ruledDetail.Path != "" {
		detail = ruledDetail
	}
	// 将文件流式传输到 w, 同时使用 io.TeeReader 来缓存到临时文件
	cachePath := s.storageCachePath(detail)
	if stor, ok := defaultService.storages[detail.Type]; ok {
		// 先检查缓存
		if cacheFile, err := osutil.OpenCache(cachePath); err == nil {
			defer cacheFile.Close()
			_, err := io.Copy(w, cacheFile)
			return err
		}
		// 从存储获取
		rc, err := stor.GetFile(ctx, detail)
		if err != nil {
			return oops.Wrapf(err, "get file from storage %s failed", detail.Type)
		}
		defer rc.Close()
		// 读取到临时文件, 避免频繁从远程存储获取文件
		cacheFile, err := osutil.CreateCache(cachePath)
		if err != nil {
			return oops.Wrapf(err, "create cache file failed")
		}
		defer func() {
			cacheFile.Close()
			if err != nil {
				osutil.RemoveNow(cacheFile.Name())
			}
		}()
		tr := io.TeeReader(rc, cacheFile)
		_, err = io.Copy(w, tr)
		return err
	}
	return oops.Errorf("storage type %s not found", detail.Type)
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

func (s *Service) StorageSaveAllSize(ctx context.Context, inputPath, storDirPath, fileName string) (*shared.StorageInfo, error) {
	if len(s.storages) == 0 {
		return nil, oops.New("no storage configured")
	}
	var originalDetail, regularDetail, thumbDetail *shared.StorageDetail
	if s.storCfg.OriginalType != "" {
		err := func() error {
			origStor := s.storages[shared.StorageType(s.storCfg.OriginalType)]
			if origStor == nil {
				return oops.Errorf("original storage type %s not found", s.storCfg.OriginalType)
			}
			file, err := os.Open(inputPath)
			if err != nil {
				return oops.Wrapf(err, "failed to open file for original storage %s", s.storCfg.OriginalType)
			}
			defer file.Close()
			origPath := path.Join(storDirPath, fileName)
			originalDetail, err = origStor.Save(ctx, file, origPath)
			if err != nil {
				return oops.Wrapf(err, "failed to save original file to storage %s", s.storCfg.OriginalType)
			}
			mimeType, err := mimetype.DetectFile(inputPath)
			if err != nil {
				return oops.Wrapf(err, "failed to detect mime type for original storage %s", s.storCfg.OriginalType)
			}
			originalDetail.Mime = mimeType.String()
			return nil
		}()
		if err != nil {
			return nil, err
		}
	}
	fileNameWithOutExt := fileName
	if ext := filepath.Ext(fileName); ext != "" {
		fileNameWithOutExt = fileName[:len(fileName)-len(ext)]
	}
	compressedPath := filepath.Join(s.storCfg.CacheDir, "compress", fmt.Sprintf("regular_%s.%s", fileNameWithOutExt, s.storCfg.RegularFormat))
	err := imgtool.Compress(inputPath, compressedPath, s.storCfg.RegularFormat, s.storCfg.RegularLength)
	if err != nil {
		return nil, oops.Wrapf(err, "failed to compress image for regular storage %s", s.storCfg.RegularType)
	}
	compressedFile, err := os.Open(compressedPath)
	if err != nil {
		return nil, oops.Wrapf(err, "failed to open compressed file for regular storage %s", s.storCfg.RegularType)
	}
	defer os.Remove(compressedFile.Name())
	defer compressedFile.Close()
	if s.storCfg.RegularType != "" {
		regStor := s.storages[shared.StorageType(s.storCfg.RegularType)]
		if regStor == nil {
			return nil, oops.Errorf("regular storage type %s not found", s.storCfg.RegularType)
		}
		regPath := path.Join(storDirPath, fmt.Sprintf("%s_regular.%s", fileNameWithOutExt, s.storCfg.RegularFormat))
		var err error
		regularDetail, err = regStor.Save(ctx, compressedFile, regPath)
		if err != nil {
			return nil, oops.Wrapf(err, "failed to save regular file to storage %s", s.storCfg.RegularType)
		}
		mimeType, err := mimetype.DetectFile(compressedPath)
		if err != nil {
			return nil, oops.Wrapf(err, "failed to detect mime type for regular storage %s", s.storCfg.RegularType)
		}
		regularDetail.Mime = mimeType.String()
	}
	compressedPath2 := filepath.Join(s.storCfg.CacheDir, "compress", fmt.Sprintf("thumb_%s.%s", fileNameWithOutExt, s.storCfg.ThumbFormat))
	err = imgtool.Compress(inputPath, compressedPath2, s.storCfg.ThumbFormat, s.storCfg.ThumbLength)
	if err != nil {
		return nil, oops.Wrapf(err, "failed to compress image for thumb storage %s", s.storCfg.ThumbType)
	}
	compressedFile2, err := os.Open(compressedPath2)
	if err != nil {
		return nil, oops.Wrapf(err, "failed to open compressed file for thumb storage %s", s.storCfg.ThumbType)
	}
	defer os.Remove(compressedFile2.Name())
	defer compressedFile2.Close()
	if s.storCfg.ThumbType != "" {
		thumbStor := s.storages[shared.StorageType(s.storCfg.ThumbType)]
		if thumbStor == nil {
			return nil, oops.Errorf("thumb storage type %s not found", s.storCfg.ThumbType)
		}
		thumbPath := path.Join(storDirPath, fmt.Sprintf("%s_thumb.%s", fileNameWithOutExt, s.storCfg.ThumbFormat))
		var err error
		thumbDetail, err = thumbStor.Save(ctx, compressedFile2, thumbPath)
		if err != nil {
			return nil, oops.Wrapf(err, "failed to save thumb file to storage %s", s.storCfg.ThumbType)
		}
		mimeType, err := mimetype.DetectFile(compressedPath2)
		if err != nil {
			return nil, oops.Wrapf(err, "failed to detect mime type for thumb storage %s", s.storCfg.ThumbType)
		}
		thumbDetail.Mime = mimeType.String()
	}
	return &shared.StorageInfo{
		Original: originalDetail,
		Regular:  regularDetail,
		Thumb:    thumbDetail,
	}, nil
}

func (s *Service) StorageSaveOriginal(ctx context.Context, file io.Reader, storDirPath, fileName string) (*shared.StorageDetail, error) {
	if len(s.storages) == 0 {
		return nil, oops.New("no storage configured")
	}
	if s.storCfg.OriginalType == "" {
		return nil, oops.New("no original storage configured")
	}
	origStor := s.storages[shared.StorageType(s.storCfg.OriginalType)]
	if origStor == nil {
		return nil, oops.Errorf("original storage type %s not found", s.storCfg.OriginalType)
	}
	origPath := path.Join(storDirPath, fileName)
	originalDetail, err := origStor.Save(ctx, file, origPath)
	if err != nil {
		return nil, oops.Wrapf(err, "failed to save original file to storage %s", s.storCfg.OriginalType)
	}
	mimeType, err := mimetype.DetectFile(fileName)
	if err != nil {
		return nil, oops.Wrapf(err, "failed to detect mime type for original storage %s", s.storCfg.OriginalType)
	}
	originalDetail.Mime = mimeType.String()
	return originalDetail, nil
}
