package imgtool

import (
	"fmt"
	"image"
	"os"
	"os/exec"
	"runtime"

	"github.com/gen2brain/avif"
	"github.com/krau/ManyACG/internal/infra/config"
	"github.com/krau/ManyACG/types"
)

var (
	ffmpegAvailable bool
	vipsFormat      map[string]struct{}
	nativeFormat    = map[string]struct{}{"jpeg": {}, "jpg": {}, "png": {}, "webp": {}, "avif": {}}
)

func Init() {
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

func GetImageSize(img image.Image) (int, int, error) {
	if img == nil {
		return 0, 0, fmt.Errorf("nil image")
	}
	bounds := img.Bounds()
	if bounds.Empty() {
		return 0, 0, fmt.Errorf("empty image")
	}
	return bounds.Dx(), bounds.Dy(), nil
}

func CompressImage(inputPath, outputPath, format string, maxEdgeLength int) error {
	if _, ok := vipsFormat[format]; ok {
		fmt.Printf("Using vips to compress image: %s , %s , %s\n", inputPath, outputPath, format)
		return compressImageVIPS(inputPath, outputPath, format, maxEdgeLength)
	}
	if ffmpegAvailable {
		fmt.Printf("Using ffmpeg to compress image: %s , %s , %s\n", inputPath, outputPath, format)
		return compressImageByFFmpeg(inputPath, outputPath, maxEdgeLength)
	}
	if _, ok := nativeFormat[format]; ok {
		fmt.Printf("Using native to compress image: %s , %s , %s\n", inputPath, outputPath, format)
		return compressImageNative(inputPath, outputPath, format, maxEdgeLength)
	}
	return fmt.Errorf("unsupported image format: %s", format)
}

func CompressImageForTelegram(input []byte) ([]byte, error) {
	if _, ok := vipsFormat["jpeg"]; ok {
		return compressImageForTelegramByVIPS(input)
	}
	tmpFile, err := os.CreateTemp(config.Get().Storage.CacheDir, "imgtool_*.png")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	distFile, err := os.CreateTemp(config.Get().Storage.CacheDir, "imgtool_*.jpg")
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
		err = compressImageByFFmpeg(tmpFile.Name(), distFile.Name(), types.RegularPhotoSideLength)
		if err != nil {
			return nil, fmt.Errorf("failed to compress image by ffmpeg: %w", err)
		}
		result, err := os.ReadFile(distFile.Name())
		if err != nil {
			return nil, fmt.Errorf("failed to read temp file: %w", err)
		}
		return result, nil
	}
	err = compressImageNative(tmpFile.Name(), distFile.Name(), "jpeg", types.RegularPhotoSideLength)
	if err != nil {
		return nil, fmt.Errorf("failed to compress image natively: %w", err)
	}
	result, err := os.ReadFile(distFile.Name())
	if err != nil {
		return nil, fmt.Errorf("failed to read temp file: %w", err)
	}
	return result, nil
}
