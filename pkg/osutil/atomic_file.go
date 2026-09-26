package osutil

import (
	"math/rand/v2"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// siblingPath 在 target 同目录下生成新文件名, 保留 target 的扩展名
// (ffmpeg 等工具依赖扩展名判断格式).
func siblingPath(target, suffix string) string {
	ext := filepath.Ext(target)
	return filepath.Join(filepath.Dir(target), strings.TrimSuffix(filepath.Base(target), ext)+suffix+ext)
}

// CreateTempSibling 在 target 所在目录创建临时文件,
// 用于先写临时文件, 完成后再重命名到 target 的场景.
func CreateTempSibling(target string) (*os.File, error) {
	return os.CreateTemp(filepath.Dir(target), filepath.Base(siblingPath(target, ".tmp-*")))
}

// UniquePath 返回 target 同目录下带随机后缀的文件名.
// 用于只写一次的临时产物: 每次调用得到不同的路径, 并发写入互不影响.
func UniquePath(target string) string {
	return siblingPath(target, "-"+strconv.FormatUint(rand.Uint64(), 36))
}

// CommitTempFile 将临时文件重命名为 target, 使其以完整内容出现在目标路径上.
func CommitTempFile(tmpPath, target string) error {
	// 临时文件的权限受 umask 影响, 显式设置以与其他缓存文件一致;
	// 设置失败不会影响文件内容, 忽略即可.
	_ = os.Chmod(tmpPath, 0o644)
	return os.Rename(tmpPath, target)
}
