//go:build vips

package mediatool

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/cshum/vipsgen/vips817"
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
	if maxEdgeLength > 0 {
		if width > height {
			if width > maxEdgeLength {
				scale = float64(maxEdgeLength) / float64(width)
			}
		} else {
			if height > maxEdgeLength {
				scale = float64(maxEdgeLength) / float64(height)
			}
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
		// vipsgen defaults to AVC; AVIF requires the AV1 codec.
		heifOptions := vips.DefaultHeifsaveOptions()
		heifOptions.Compression = vips.HeifCompressionAv1
		err = img.Heifsave(outputPath, heifOptions)
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
	plan := planTelegramPhoto(img.Width(), img.Height(), 0)
	if !plan.needsTransform() && len(input) <= TelegramMaxPhotoFileSize {
		return input, nil
	}
	if err := applyTelegramPhotoPlan(img, plan); err != nil {
		return nil, err
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
	defer img.Close()
	if err := os.MkdirAll(filepath.Dir(outputPath), os.ModePerm); err != nil {
		return err
	}
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return fmt.Errorf("failed to get file info: %w", err)
	}
	plan := planTelegramPhoto(img.Width(), img.Height(), 0)
	if !plan.needsTransform() && fileInfo.Size() <= int64(TelegramMaxPhotoFileSize) {
		return fileutil.CopyFile(filePath, outputPath)
	}
	if err := applyTelegramPhotoPlan(img, plan); err != nil {
		return err
	}
	if err := img.Jpegsave(outputPath, vips.DefaultJpegsaveOptions()); err != nil {
		return fmt.Errorf("failed to save image to file: %w", err)
	}
	return nil
}

// applyTelegramPhotoPlan 缩放图片并居中加边, 使其成为合法的 Telegram photo.
func applyTelegramPhotoPlan(img *vips.Image, plan telegramPhotoPlan) error {
	if !plan.needsTransform() {
		return nil
	}
	if plan.scale < 1 {
		if err := img.Resize(plan.scale, vips.DefaultResizeOptions()); err != nil {
			return fmt.Errorf("failed to resize image: %w", err)
		}
	}
	// 缩放取整后可能比画布大 1px
	if img.Width() > plan.canvasWidth || img.Height() > plan.canvasHeight {
		left := (img.Width() - plan.canvasWidth) / 2
		top := (img.Height() - plan.canvasHeight) / 2
		if err := img.ExtractArea(left, top, plan.canvasWidth, plan.canvasHeight); err != nil {
			return fmt.Errorf("failed to crop image: %w", err)
		}
	}
	if img.Width() == plan.canvasWidth && img.Height() == plan.canvasHeight {
		return nil
	}
	left := (plan.canvasWidth - img.Width()) / 2
	top := (plan.canvasHeight - img.Height()) / 2
	if err := img.Embed(left, top, plan.canvasWidth, plan.canvasHeight, &vips.EmbedOptions{Extend: vips.ExtendWhite}); err != nil {
		return fmt.Errorf("failed to embed image: %w", err)
	}
	return nil
}
