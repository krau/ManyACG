package imgtool

import (
	"fmt"
	"image"
	"os"
	"os/exec"
	"runtime"

	"github.com/cshum/vipsgen/vips"
	"github.com/gen2brain/avif"
	"github.com/krau/ManyACG/types"
	ffmpeg "github.com/u2takey/ffmpeg-go"
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

// 使用 ffmpeg 压缩图片
func compressImageByFFmpeg(inputPath, outputPath string, maxEdgeLength int) error {
	file, err := os.Open(inputPath)
	if err != nil {
		return err
	}
	defer file.Close()
	img, _, err := image.DecodeConfig(file)
	if err != nil {
		return err
	}
	var vfKwArg ffmpeg.KwArgs
	if img.Width > int(maxEdgeLength) || img.Height > int(maxEdgeLength) {
		if img.Width > img.Height {
			vfKwArg = ffmpeg.KwArgs{"vf": fmt.Sprintf("scale=%d:-1:flags=lanczos", maxEdgeLength)}
		} else {
			vfKwArg = ffmpeg.KwArgs{"vf": fmt.Sprintf("scale=-1:%d:flags=lanczos", maxEdgeLength)}
		}
	}
	if err := ffmpeg.Input(inputPath).Output(outputPath, vfKwArg).OverWriteOutput().Run(); err != nil {
		return fmt.Errorf("failed to compress image: %w", err)
	}
	return nil
}

func CompressImage(inputPath, outputPath, format string, maxEdgeLength int) error {
	img, err := vips.NewImageFromFile(inputPath, vips.DefaultLoadOptions())
	if err != nil {
		return fmt.Errorf("failed to create image from file: %w", err)
	}
	defer img.Close()
	width := img.Width()
	height := img.Height()

	var scale float64 = 1.0
	if width > height {
		if width > maxEdgeLength {
			scale = float64(maxEdgeLength) / float64(width)
		}
	} else {
		if height > maxEdgeLength {
			scale = float64(maxEdgeLength) / float64(height)
		}
	}
	if scale < 1.0 {
		err = img.Resize(scale, vips.DefaultResizeOptions())
		if err != nil {
			return fmt.Errorf("failed to resize image: %w", err)
		}
	}
	if _, ok := vipsFormat[format]; !ok {
		if ffmpegAvailable {
			return compressImageByFFmpeg(inputPath, outputPath, maxEdgeLength)
		}
		if _, ok := nativeFormat[format]; ok {
			return compressImageNative(inputPath, outputPath, format, maxEdgeLength)
		}
		return fmt.Errorf("unsupported image format: %s", format)
	}
	switch format {
	case "jpeg", "jpg":
		err = img.Jpegsave(outputPath, vips.DefaultJpegsaveOptions())
	case "png":
		err = img.Pngsave(outputPath, vips.DefaultPngsaveOptions())
	case "webp":
		err = img.Webpsave(outputPath, vips.DefaultWebpsaveOptions())
	// case "avif":
	// 	err = img.Heifsave(outputPath, vips.DefaultHeifsaveOptions())
	default:
		return fmt.Errorf("unsupported image format: %s", format)
	}
	if err != nil {
		return fmt.Errorf("failed to save image to file: %w", err)
	}
	return nil
}

func CompressImageForTelegram(input []byte) ([]byte, error) {
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
