//go:build linux && amd64 && !without_vips

package imgtool

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/cshum/vipsgen/vips"
	"github.com/duke-git/lancet/v2/fileutil"
)

func init() {
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
	if vips.HasOperation("heifsave") {
		vipsFormat["avif"] = struct{}{}
	}
}

func compressImageVIPS(inputPath, outputPath, format string, maxEdgeLength int) error {
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
	switch format {
	case "jpeg", "jpg":
		err = img.Jpegsave(outputPath, vips.DefaultJpegsaveOptions())
	case "png":
		err = img.Pngsave(outputPath, vips.DefaultPngsaveOptions())
	case "webp":
		err = img.Webpsave(outputPath, vips.DefaultWebpsaveOptions())
	case "avif":
		err = img.Heifsave(outputPath, vips.DefaultHeifsaveOptions())
	default:
		return fmt.Errorf("unsupported image format: %s", format)
	}
	if err != nil {
		return fmt.Errorf("failed to save image to file: %w", err)
	}
	return nil
}

func compressImageForTelegramByVIPS(input []byte) ([]byte, error) {
	img, err := vips.NewImageFromBuffer(input, vips.DefaultLoadOptions())
	if err != nil {
		return nil, fmt.Errorf("failed to create image from buffer: %w", err)
	}
	defer img.Close()
	width := img.Width()
	height := img.Height()
	inputLen := len(input)
	currentTotalSideLength := width + height
	if currentTotalSideLength <= TelegramMaxPhotoTotalSideLength &&
		inputLen <= TelegramMaxPhotoFileSize {
		return input, nil
	}
	var scale float64 = 1.0
	if currentTotalSideLength > TelegramMaxPhotoTotalSideLength {
		scale = float64(TelegramMaxPhotoTotalSideLength) / float64(currentTotalSideLength)
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

func compressImageForTelegramByVIPSFromFile(filePath, outputPath string) error {
	img, err := vips.NewImageFromFile(filePath, vips.DefaultLoadOptions())
	if err != nil {
		return fmt.Errorf("failed to create image from file: %w", err)
	}
	if os.MkdirAll(filepath.Dir(outputPath), os.ModePerm); err != nil {
		return err
	}
	defer img.Close()
	width := img.Width()
	height := img.Height()
	inputLen, err := os.Stat(filePath)
	if err != nil {
		return fmt.Errorf("failed to get file info: %w", err)
	}
	currentTotalSideLength := width + height
	if currentTotalSideLength <= TelegramMaxPhotoTotalSideLength &&
		inputLen.Size() <= int64(TelegramMaxPhotoFileSize) {
		return fileutil.CopyFile(filePath, outputPath)
	}
	var scale float64 = 1.0
	if currentTotalSideLength > TelegramMaxPhotoTotalSideLength {
		scale = float64(TelegramMaxPhotoTotalSideLength) / float64(currentTotalSideLength)
	}
	if scale < 1.0 {
		err = img.Resize(scale, vips.DefaultResizeOptions())
		if err != nil {
			return fmt.Errorf("failed to resize image: %w", err)
		}
	}
	err = img.Jpegsave(outputPath, vips.DefaultJpegsaveOptions())
	if err != nil {
		return fmt.Errorf("failed to save image to buffer: %w", err)
	}
	return nil
}
