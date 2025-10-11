package osutil

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/duke-git/lancet/v2/strutil"
	"github.com/duke-git/lancet/v2/validator"
)

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

var fileNameReplacer = strings.NewReplacer(
	" ", "_",
	"/", "_",
	"\\", "_",
	":", "_",
	"*", "_",
	"?", "_",
	"\"", "_",
	"<", "_",
	">", "_",
	"|", "_",
	"%", "_",
	"#", "_",
	"+", "_",
	"'", "_",
	"`", "_",
	"\t", "_",
	"\r", "_",
	"\n", "_",
)

func SanitizeFileName(fileName string) string {
	fname := strutil.RemoveWhiteSpace(fileNameReplacer.Replace(fileName), true)
	fname = strings.Map(func(r rune) rune {
		if r < 0x20 || r == 0x7F {
			return '_'
		}
		if validator.IsPrintable(string(r)) {
			return r
		}
		return '_'
	}, fname)
	return fname
}
