package imgtool

import (
	"fmt"

	"github.com/cshum/vipsgen/vips"
)

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
