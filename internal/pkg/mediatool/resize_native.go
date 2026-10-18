package mediatool

import (
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"os"

	"github.com/HugoSmits86/nativewebp"
	"github.com/gen2brain/avif"
	"golang.org/x/image/draw"
	_ "golang.org/x/image/webp"
)

func compressImageForTelegramNative(inputPath, outputPath string) error {
	imgFile, err := os.Open(inputPath)
	if err != nil {
		return fmt.Errorf("failed to open input file: %w", err)
	}
	defer imgFile.Close()

	srcImg, _, err := image.Decode(imgFile)
	if err != nil {
		return fmt.Errorf("failed to decode image: %w", err)
	}
	width, height, err := GetImgSize(srcImg)
	if err != nil {
		return err
	}
	plan := planTelegramPhoto(width, height, TelegramMaxPhotoSideLength)
	dstImg := telegramPhotoImage(srcImg, plan)

	outFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outFile.Close()

	if err := jpeg.Encode(outFile, dstImg, &jpeg.Options{Quality: 85}); err != nil {
		return fmt.Errorf("failed to encode JPEG: %w", err)
	}
	return nil
}

// telegramPhotoImage 将图片等比缩放并居中放置到加边后的画布上, 无需变换时返回原图.
func telegramPhotoImage(src image.Image, plan telegramPhotoPlan) image.Image {
	if !plan.needsTransform() {
		return src
	}
	canvas := image.NewRGBA(image.Rect(0, 0, plan.canvasWidth, plan.canvasHeight))
	draw.Draw(canvas, canvas.Bounds(), image.NewUniform(color.White), image.Point{}, draw.Src)
	left := (plan.canvasWidth - plan.width) / 2
	top := (plan.canvasHeight - plan.height) / 2
	target := image.Rect(left, top, left+plan.width, top+plan.height)
	// catmullrom , very slow but good quality
	draw.CatmullRom.Scale(canvas, target, src, src.Bounds(), draw.Over, nil)
	return canvas
}

func compressImageNative(inputPath, outputPath, format string, maxEdgeLength int) error {
	imgFile, err := os.Open(inputPath)
	if err != nil {
		return fmt.Errorf("failed to open input file: %w", err)
	}
	defer imgFile.Close()

	srcImg, _, err := image.Decode(imgFile)
	if err != nil {
		return fmt.Errorf("failed to decode image: %w", err)
	}

	width := srcImg.Bounds().Dx()
	height := srcImg.Bounds().Dy()

	scale := 1.0
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
	var dstImg image.Image
	if scale < 1.0 {
		newWidth := int(float64(width) * scale)
		newHeight := int(float64(height) * scale)
		dst := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))
		// catmullrom , very slow but good quality
		draw.CatmullRom.Scale(dst, dst.Bounds(), srcImg, srcImg.Bounds(), draw.Over, nil)
		dstImg = dst
	} else {
		dstImg = srcImg
	}

	outFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outFile.Close()

	switch format {
	case "jpeg", "jpg":
		opts := &jpeg.Options{Quality: 85}
		if err := jpeg.Encode(outFile, dstImg, opts); err != nil {
			return fmt.Errorf("failed to encode JPEG: %w", err)
		}
	case "png":
		encoder := &png.Encoder{CompressionLevel: png.BestCompression}
		if err := encoder.Encode(outFile, dstImg); err != nil {
			return fmt.Errorf("failed to encode PNG: %w", err)
		}
	case "webp":
		if err := nativewebp.Encode(outFile, dstImg, nil); err != nil {
			return fmt.Errorf("failed to encode WebP: %w", err)
		}
	case "avif":
		if err := avif.Encode(outFile, dstImg); err != nil {
			return fmt.Errorf("failed to encode AVIF: %w", err)
		}
	default:
		return fmt.Errorf("unsupported image format: %s", format)
	}

	return nil
}
