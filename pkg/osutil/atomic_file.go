package osutil

import (
	"io/fs"
	"math/rand/v2"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// TempFilePrefix 是 CreateTempSibling 生成的临时文件名在扩展名前插入的前缀,
// 便于识别与清理残留的临时文件.
const TempFilePrefix = ".tmp-"

// siblingPath 在 target 同目录下生成新文件名, 保留 target 的扩展名
// (ffmpeg 等工具依赖扩展名判断格式).
func siblingPath(target, suffix string) string {
	ext := filepath.Ext(target)
	return filepath.Join(filepath.Dir(target), strings.TrimSuffix(filepath.Base(target), ext)+suffix+ext)
}

// IsTempFile 判断文件名是否为临时文件
func IsTempFile(name string) bool {
	return strings.Contains(filepath.Base(name), TempFilePrefix)
}

// CreateTempSibling 在 target 所在目录创建临时文件,
// 用于先写临时文件, 完成后再重命名到 target 的场景.
func CreateTempSibling(target string) (*os.File, error) {
	return os.CreateTemp(filepath.Dir(target), filepath.Base(siblingPath(target, TempFilePrefix+"*")))
}

// TempSiblingPath 返回 target 同目录下的临时文件路径, 扩展名保留在末尾.
// 供 vips/ffmpeg 等自行创建输出文件的场景使用, 便于统一清理残留.
func TempSiblingPath(target string) string {
	return siblingPath(target, TempFilePrefix+strconv.FormatUint(rand.Uint64(), 36))
}

// UniquePath 返回 target 同目录下带随机后缀的文件名.
// 用于只写一次的产物: 每次调用得到不同的路径, 并发写入互不影响.
func UniquePath(target string) string {
	return siblingPath(target, "-"+strconv.FormatUint(rand.Uint64(), 36))
}

// CommitTempFile 将临时文件重命名为 target, 使其以完整内容出现在目标路径上.
//
// 保证范围限于进程层面: rename 是原子的, 其他读者不会读到写入中的内容, 进程被结束时
// 目标路径也不会留下半成品. 目标内容不保证在掉电时已落盘(缓存内容可再生, 不做 fsync).
func CommitTempFile(tmpPath, target string) error {
	// 临时文件的权限受 umask 影响, 显式设置以与其他缓存文件一致;
	// 设置失败不会影响文件内容, 忽略即可.
	_ = os.Chmod(tmpPath, 0o644)
	return os.Rename(tmpPath, target)
}

// RemoveStaleTempFiles 删除 dir 下修改时间早于 olderThan 的临时文件,
// 用于清理进程被强制结束等原因残留的临时文件, 返回删除数量.
func RemoveStaleTempFiles(dir string, olderThan time.Duration) (int, error) {
	if _, err := os.Stat(dir); err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, err
	}
	deadline := time.Now().Add(-olderThan)
	removed := 0
	var firstErr error
	walkErr := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			if firstErr == nil {
				firstErr = err
			}
			return nil
		}
		if d.IsDir() || !IsTempFile(d.Name()) {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			if firstErr == nil {
				firstErr = err
			}
			return nil
		}
		// 仍可能在被写入的临时文件不删除
		if info.ModTime().After(deadline) {
			return nil
		}
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			if firstErr == nil {
				firstErr = err
			}
			return nil
		}
		removed++
		return nil
	})
	if walkErr != nil {
		return removed, walkErr
	}
	return removed, firstErr
}
