//go:build linux && amd64

package imgtool

import (
	"fmt"

	"github.com/cshum/vipsgen/vips"
	"github.com/krau/ManyACG/types"
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
	// if vips.HasOperation("heifsave") {
	// 	vipsFormat["avif"] = struct{}{}
	// }
	// https://github.com/cshum/vipsgen/issues/58
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
