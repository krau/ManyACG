package osutil

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"sync"
	"time"
)

var (
	cacheMu   sync.Mutex
	refCounts = map[string]int{}           // 文件引用计数
	timers    = map[string]*time.Timer{}   // 延迟删除定时器
	fileTTLs  = map[string]time.Duration{} // 每个文件的 TTL（可选）

	cachettl      = 10 * time.Minute // 全局默认 TTL
	onRemoveError func(path string, err error)
)

// OnRemoveError 是一个可选回调函数，在删除文件失败时调用

// SetCacheTTL 设置全局文件缓存延迟删除时间
func SetCacheTTL(d time.Duration) {
	cacheMu.Lock()
	cachettl = d
	cacheMu.Unlock()
}

func SetOnRemoveError(f func(path string, err error)) {
	cacheMu.Lock()
	onRemoveError = f
	cacheMu.Unlock()
}

// OpenCache 打开一个文件并增加引用计数。
// 使用全局默认 TTL。
func OpenCache(path string) (*File, error) {
	return OpenCacheWithTTL(path, cachettl)
}

func CreateCache(path string) (*File, error) {
	if err := os.MkdirAll(filepath.Dir(path), os.ModePerm); err != nil {
		return nil, err
	}
	f, err := os.Create(path)
	if err != nil {
		return nil, err
	}

	cacheMu.Lock()
	defer cacheMu.Unlock()

	refCounts[path]++
	fileTTLs[path] = cachettl

	// 如果存在删除定时器，停止并删除
	if t, ok := timers[path]; ok {
		t.Stop()
		delete(timers, path)
	}

	return &File{File: f, path: path}, nil
}

// OpenCacheWithTTL 打开一个文件并增加引用计数，使用自定义 TTL。
func OpenCacheWithTTL(path string, ttl time.Duration) (*File, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	cacheMu.Lock()
	defer cacheMu.Unlock()

	refCounts[path]++
	fileTTLs[path] = ttl

	// 如果存在删除定时器，停止并删除
	if t, ok := timers[path]; ok {
		t.Stop()
		delete(timers, path)
	}

	return &File{File: f, path: path}, nil
}

// MkCache 创建文件并自动加入缓存管理
func MkCache(path string, data []byte) (*File, error) {
	if err := MkFile(path, data); err != nil {
		return nil, err
	}
	return OpenCache(path)
}

// RemoveNow 立即删除文件（无论是否被引用）
func RemoveNow(path string) error {
	cacheMu.Lock()
	if t, ok := timers[path]; ok {
		t.Stop()
		delete(timers, path)
	}
	delete(refCounts, path)
	delete(fileTTLs, path)
	cacheMu.Unlock()

	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		if onRemoveError != nil {
			onRemoveError(path, err)
		}
		return err
	}
	return nil
}

// markClosed 在最后一个 Close 调用时启动延迟删除
func markClosed(path string) {
	cacheMu.Lock()
	defer cacheMu.Unlock()

	cnt, ok := refCounts[path]
	if !ok {
		return
	}

	cnt--
	if cnt > 0 {
		refCounts[path] = cnt
		return
	}

	// 没有引用了，删除引用计数
	delete(refCounts, path)
	ttl := fileTTLs[path]
	delete(fileTTLs, path)

	// 启动延迟删除定时器
	timers[path] = time.AfterFunc(ttl, func() {
		cacheMu.Lock()
		defer cacheMu.Unlock()

		// 文件在定时期间可能被重新引用，若如此则不删除
		if _, stillUsed := refCounts[path]; stillUsed {
			return
		}

		// 执行删除
		err := os.Remove(path)
		if err != nil && !os.IsNotExist(err) && onRemoveError != nil {
			onRemoveError(path, err)
		}
		delete(timers, path)
	})
}

func MkFile(path string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), os.ModePerm); err != nil {
		return err
	}

	tmp := fmt.Sprintf("%s.tmp.%d.%d", path, os.Getpid(), rand.Int())
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		_ = os.Remove(tmp)
		return err
	}

	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return err
	}

	return nil
}

type File struct {
	*os.File
	path string
}

func (f *File) Close() error {
	err := f.File.Close()
	markClosed(f.path)
	return err
}

type TempFile struct {
	*os.File
}

func (t *TempFile) Close() error {
	err := t.File.Close()
	if err != nil {
		return err
	}
	return os.Remove(t.File.Name())
}

func OpenTemp(path string) (*TempFile, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	return &TempFile{f}, nil
}
