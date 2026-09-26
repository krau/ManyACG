package mediatool

import (
	"fmt"
	"image"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/krau/ManyACG/internal/infra/config/runtimecfg"
	"github.com/krau/ManyACG/pkg/log"
	"github.com/krau/ManyACG/pkg/osutil"
	"github.com/krau/ManyACG/pkg/strutil"
)

var (
	ffmpegAvailable bool
	vipsFormat      map[string]struct{}
	nativeFormat    = map[string]struct{}{"jpeg": {}, "jpg": {}, "png": {}, "webp": {}, "avif": {}}
)

func init() {
	switch runtime.GOOS {
	case "windows":
		_, err := exec.LookPath("ffmpeg.exe")
		if err == nil {
			ffmpegAvailable = true
		}
	default:
		_, err := exec.LookPath("ffmpeg")
		if err == nil {
			ffmpegAvailable = true
		}
	}
}

func FFmpegAvailable() bool {
	return ffmpegAvailable
}

func GetImgSize(img image.Image) (int, int, error) {
	if img == nil {
		return 0, 0, fmt.Errorf("nil image")
	}
	bounds := img.Bounds()
	if bounds.Empty() {
		return 0, 0, fmt.Errorf("empty image")
	}
	return bounds.Dx(), bounds.Dy(), nil
}

func GetImgSizeFromReader(r io.Reader) (int, int, error) {
	img, _, err := image.Decode(r)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to decode image: %w", err)
	}
	return GetImgSize(img)
}

// CompressImg compresses the image at inputPath and writes it to a file derived
// from outputPath (a unique suffix is appended to avoid concurrent writers), then
// returns the actual output path.
//
// The input image will be resized so that its longest edge does not exceed maxEdgeLength,
//
// If the maxEdgeLength <= 0, no resizing will be performed.
func CompressImg(inputPath, outputPath, format string, maxEdgeLength int) (string, error) {
	if err := os.MkdirAll(filepath.Dir(outputPath), os.ModePerm); err != nil {
		return "", err
	}
	// 先写临时文件, 成功后再提交到唯一路径:
	// 压缩失败时只留下可被清理的临时文件, 并发调用也不会写到同一个文件
	tmpFile, err := osutil.CreateTempSibling(outputPath)
	if err != nil {
		return "", err
	}
	tmpPath := tmpFile.Name()
	if err := tmpFile.Close(); err != nil {
		os.Remove(tmpPath)
		return "", err
	}
	defer os.Remove(tmpPath)
	if err := compressImgToFile(inputPath, tmpPath, format, maxEdgeLength); err != nil {
		return "", err
	}
	outPath := osutil.UniquePath(outputPath)
	if err := osutil.CommitTempFile(tmpPath, outPath); err != nil {
		return "", err
	}
	return outPath, nil
}

func compressImgToFile(inputPath, outputPath, format string, maxEdgeLength int) error {
	if _, ok := vipsFormat[format]; ok {
		log.Debug("compressing image", "method", "vips", "input", inputPath, "output", outputPath, "format", format)
		err := compressImageVIPS(inputPath, outputPath, format, maxEdgeLength)
		if err != nil {
			return fmt.Errorf("failed to compress image with vips: %w", err)
		}
		return nil
	}
	if ffmpegAvailable {
		log.Debug("compressing image", "method", "ffmpeg", "input", inputPath, "output", outputPath, "format", format)
		err := compressImageByFFmpeg(inputPath, outputPath, maxEdgeLength)
		if err != nil {
			return fmt.Errorf("failed to compress image with ffmpeg: %w", err)
		}
		return nil
	}
	if _, ok := nativeFormat[format]; ok {
		log.Debug("compressing image", "method", "native", "input", inputPath, "output", outputPath, "format", format)
		err := compressImageNative(inputPath, outputPath, format, maxEdgeLength)
		if err != nil {
			return fmt.Errorf("failed to compress image with native: %w", err)
		}
		return nil
	}
	return fmt.Errorf("unsupported image format: %s", format)
}

func CompressImgForTelegram(input []byte) ([]byte, error) {
	if _, ok := vipsFormat["jpeg"]; ok {
		return compressImageForTelegramByVIPS(input)
	}
	cacheDir := runtimecfg.Get().Storage.CacheDir
	tmpFile, err := osutil.CreateTempSibling(filepath.Join(cacheDir, "mediatool_telegram.png"))
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	distFile, err := osutil.CreateTempSibling(filepath.Join(cacheDir, "mediatool_telegram.jpg"))
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(distFile.Name())
	defer distFile.Close()

	err = os.WriteFile(tmpFile.Name(), input, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to write temp file: %w", err)
	}
	if ffmpegAvailable {
		err = compressImageByFFmpeg(tmpFile.Name(), distFile.Name(), TelegramMaxPhotoSideLength)
		if err != nil {
			return nil, fmt.Errorf("failed to compress image by ffmpeg: %w", err)
		}
		result, err := os.ReadFile(distFile.Name())
		if err != nil {
			return nil, fmt.Errorf("failed to read temp file: %w", err)
		}
		return result, nil
	}
	err = compressImageNative(tmpFile.Name(), distFile.Name(), "jpeg", TelegramMaxPhotoSideLength)
	if err != nil {
		return nil, fmt.Errorf("failed to compress image natively: %w", err)
	}
	result, err := os.ReadFile(distFile.Name())
	if err != nil {
		return nil, fmt.Errorf("failed to read temp file: %w", err)
	}
	return result, nil
}

func CompressImgForTelegramFromFile(filePath string) (*osutil.TempFile, error) {
	outputPath := osutil.TempSiblingPath(filepath.Join(runtimecfg.Get().Storage.CacheDir, "compress", fmt.Sprintf("tg_%s.jpg", strutil.MD5Hash(filePath))))
	if _, ok := vipsFormat["jpeg"]; ok {
		err := compressImageForTelegramByVIPSFromFile(filePath, outputPath)
		if err != nil {
			return nil, err
		}
		f, err := os.Open(outputPath)
		if err != nil {
			return nil, err
		}
		return &osutil.TempFile{File: f}, nil
	}
	if ffmpegAvailable {
		err := compressImageByFFmpeg(filePath, outputPath, TelegramMaxPhotoSideLength)
		if err != nil {
			return nil, err
		}
		f, err := os.Open(outputPath)
		if err != nil {
			return nil, err
		}
		return &osutil.TempFile{File: f}, nil
	}
	err := compressImageNative(filePath, outputPath, "jpeg", TelegramMaxPhotoSideLength)
	if err != nil {
		return nil, err
	}
	f, err := os.Open(outputPath)
	if err != nil {
		return nil, err
	}
	return &osutil.TempFile{File: f}, nil
}
