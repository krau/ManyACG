package service

import (
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"

	"github.com/krau/ManyACG/internal/pkg/imgtool"
	"github.com/krau/ManyACG/internal/shared"
	"github.com/krau/ManyACG/pkg/osutil"
	"github.com/samber/oops"
)

func (s *Service) StorageGetFile(ctx context.Context, detail shared.StorageDetail) (*osutil.File, error) {
	if stor, ok := s.storages[detail.Type]; ok {
		// 先检查缓存
		if cacheFile, err := osutil.OpenCache(filepath.Join(s.storCfg.CacheDir, "storage", detail.Hash())); err == nil {
			return cacheFile, nil
		}
		// 从存储获取
		rc, err := stor.GetFile(ctx, detail)
		if err != nil {
			return nil, oops.Wrapf(err, "get file from storage %s failed", detail.Type)
		}
		defer rc.Close()
		// 读取到临时文件, 避免频繁从远程存储获取文件
		cacheFile, err := osutil.CreateCache(filepath.Join(s.storCfg.CacheDir, "storage", detail.Hash()))
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
	return originalDetail, nil
}
