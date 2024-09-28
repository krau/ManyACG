package common

import (
	"errors"
	"os"
	"path/filepath"
	"time"

	. "github.com/krau/ManyACG/logger"
)

// 创建文件, 自动创建目录
func MkFile(path string, data []byte) error {
	err := os.MkdirAll(filepath.Dir(path), os.ModePerm)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, os.ModePerm)
}

func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// 删除文件, 并清理空目录. 如果文件不存在则返回 nil
func PurgeFile(path string) error {
	if err := os.Remove(path); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return err
		}
	}
	return RemoveEmptyDirectories(filepath.Dir(path))
}

// 递归删除空目录
func RemoveEmptyDirectories(dirPath string) error {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return err
	}
	if len(entries) == 0 {
		err := os.Remove(dirPath)
		if err != nil {
			return err
		}
		return RemoveEmptyDirectories(filepath.Dir(dirPath))
	}
	return nil
}

// 在指定时间后删除和清理文件 (定时器)
func PurgeFileAfter(path string, td time.Duration) {
	_, err := os.Stat(path)
	if err != nil {
		Logger.Errorf("Failed to create timer for %s: %s", path, err)
		return
	}
	Logger.Debugf("Purge file after %s: %s", td, path)
	time.AfterFunc(td, func() {
		PurgeFile(path)
	})
}

func MkCache(path string, data []byte, td time.Duration) {
	if err := MkFile(path, data); err != nil {
		Logger.Errorf("failed to save cache file: %s", err)
	} else {
		go PurgeFileAfter(path, td)
	}
}
