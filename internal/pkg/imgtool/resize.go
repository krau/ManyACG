package imgtool

import (
	"fmt"
	"image"
	"io"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/gen2brain/avif"
	"github.com/krau/ManyACG/internal/infra/config/runtimecfg"
	"github.com/krau/ManyACG/pkg/log"
	"github.com/krau/ManyACG/pkg/strutil"
)

var (
	ffmpegAvailable bool
	vipsFormat      map[string]struct{}
	nativeFormat    = map[string]struct{}{"jpeg": {}, "jpg": {}, "png": {}, "webp": {}, "avif": {}}
)

func init() {
	avif.InitEncoder()
	avif.InitDecoder()
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

func GetSize(img image.Image) (int, int, error) {
	if img == nil {
		return 0, 0, fmt.Errorf("nil image")
	}
	bounds := img.Bounds()
	if bounds.Empty() {
		return 0, 0, fmt.Errorf("empty image")
	}
	return bounds.Dx(), bounds.Dy(), nil
}

func GetSizeFromReader(r io.Reader) (int, int, error) {
	img, _, err := image.Decode(r)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to decode image: %w", err)
	}
	return GetSize(img)
}

func Compress(inputPath, outputPath, format string, maxEdgeLength int) error {
	if err := os.MkdirAll(filepath.Dir(outputPath), os.ModePerm); err != nil {
		return err
	}
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

func CompressForTelegram(input []byte) ([]byte, error) {
	if _, ok := vipsFormat["jpeg"]; ok {
		return compressImageForTelegramByVIPS(input)
	}
	tmpFile, err := os.CreateTemp(runtimecfg.Get().Storage.CacheDir, "imgtool_*.png")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	distFile, err := os.CreateTemp(runtimecfg.Get().Storage.CacheDir, "imgtool_*.jpg")
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

// TempFile is a temporary file that will be deleted when closed.
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

func CompressForTelegramFromFile(filePath string) (*TempFile, error) {
	outputPath := filepath.Join(runtimecfg.Get().Storage.CacheDir, "compress", fmt.Sprintf("tg_%s_%d.jpg", strutil.MD5Hash(filePath), rand.Int()))
	if _, ok := vipsFormat["jpeg"]; ok {
		err := compressImageForTelegramByVIPSFromFile(filePath, outputPath)
		if err != nil {
			return nil, err
		}
		f, err := os.Open(outputPath)
		if err != nil {
			return nil, err
		}
		return &TempFile{f}, nil
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
		return &TempFile{f}, nil
	}
	err := compressImageNative(filePath, outputPath, "jpeg", TelegramMaxPhotoSideLength)
	if err != nil {
		return nil, err
	}
	f, err := os.Open(outputPath)
	if err != nil {
		return nil, err
	}
	return &TempFile{f}, nil
}
