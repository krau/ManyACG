package imgtool

import (
	"fmt"
	"image"
	"os/exec"
	"runtime"

	"github.com/cshum/vipsgen/vips"
	"github.com/gen2brain/avif"
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
	vips.Startup(nil)
	vipsFormat = make(map[string]struct{})
	if vips.HasOperation("jpegsave") {
		vipsFormat["jpeg"] = struct{}{}
		vipsFormat["jpg"] = struct{}{}
	}
	if vips.HasOperation("pngsave") {
		vipsFormat["png"] = struct{}{}
	}
	if vips.HasOperation("webpsave") {
		vipsFormat["webp"] = struct{}{}
	}
	// if vips.HasOperation("heifsave") {
	// 	vipsFormat["avif"] = struct{}{}
	// }
	// https://github.com/cshum/vipsgen/issues/58
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
		return compressImageVIPS(inputPath, outputPath, format, maxEdgeLength)
	}
	if ffmpegAvailable {
		return compressImageByFFmpeg(inputPath, outputPath, maxEdgeLength)
	}
	if _, ok := nativeFormat[format]; ok {
		return compressImageNative(inputPath, outputPath, format, maxEdgeLength)
	}
	return fmt.Errorf("unsupported image format: %s", format)
}

func CompressImageForTelegram(input []byte) ([]byte, error) {
	if _, ok := vipsFormat["jpeg"]; ok {
		img, err := vips.NewImageFromBuffer(input, vips.DefaultLoadOptions())
		if err != nil {
			return nil, fmt.Errorf("failed to create image from buffer: %w", err)
		}
		defer img.Close()
		width := img.Width()
		height := img.Height()
		inputLen := len(input)
		currentTotalSideLength := width + height
		if currentTotalSideLength <= types.TelegramMaxPhotoTotalSideLength &&
			inputLen <= types.TelegramMaxPhotoFileSize {
			return input, nil
		}
		var scale float64 = 1.0
		if currentTotalSideLength > types.TelegramMaxPhotoTotalSideLength {
			scale = float64(types.TelegramMaxPhotoTotalSideLength) / float64(currentTotalSideLength)
		}
		if scale < 1.0 {
			err = img.Resize(scale, vips.DefaultResizeOptions())
			if err != nil {
				return nil, fmt.Errorf("failed to resize image: %w", err)
			}
		}
		result, err := img.JpegsaveBuffer(vips.DefaultJpegsaveBufferOptions())
		if err != nil {
			return nil, fmt.Errorf("failed to save image to buffer: %w", err)
		}
		return result, nil
	}
	return nil, fmt.Errorf("vips not support jpeg")
}
