package common

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	_ "image/gif"
	"image/jpeg"
	_ "image/png"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/krau/ManyACG/config"

	"golang.org/x/image/draw"

	"sync"

	"github.com/corona10/goimagehash"
	ffmpeg "github.com/u2takey/ffmpeg-go"
)

var imageBufferPool = sync.Pool{
	New: func() interface{} {
		return new(bytes.Buffer)
	},
}

func GetImagePhash(r io.Reader) (string, error) {
	img, _, err := image.Decode(r)
	if err != nil {
		return "", err
	}
	hash, err := goimagehash.PerceptionHash(img)
	if err != nil {
		return "", err
	}
	return hash.ToString(), nil
}

func GetImageBlurScore(r io.Reader) (float64, error) {
	img, _, err := image.Decode(r)
	if err != nil {
		return 0, err
	}
	bounds := img.Bounds()
	grayImg := image.NewGray(bounds)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			grayImg.Set(x, y, img.At(x, y))
		}
	}
	bounds = grayImg.Bounds()
	laplaceImg := image.NewGray(bounds)
	for y := bounds.Min.Y + 1; y < bounds.Max.Y-1; y++ {
		for x := bounds.Min.X + 1; x < bounds.Max.X-1; x++ {
			sum := int(grayImg.GrayAt(x, y).
				Y*9 - grayImg.GrayAt(x+1, y).
				Y - grayImg.GrayAt(x-1, y).
				Y - grayImg.GrayAt(x, y+1).
				Y - grayImg.GrayAt(x, y-1).
				Y - grayImg.GrayAt(x+1, y+1).
				Y - grayImg.GrayAt(x-1, y+1).
				Y - grayImg.GrayAt(x+1, y-1).
				Y - grayImg.GrayAt(x-1, y-1).Y)
			laplaceImg.SetGray(x, y, color.Gray{uint8(sum / 8)})
		}
	}

	mean := 0.0
	variance := 0.0
	pixelCount := 0

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			grayValue := float64(laplaceImg.GrayAt(x, y).Y)
			mean += grayValue
			pixelCount++

		}
	}

	mean /= float64(pixelCount)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			grayValue := float64(laplaceImg.GrayAt(x, y).Y)
			variance += (grayValue - mean) * (grayValue - mean)
		}
	}
	variance /= float64(pixelCount)

	return strconv.ParseFloat(strconv.FormatFloat(variance, 'f', 2, 64), 64)
}

// ResizeImage resizes an image to the specified width and height.
//
// It use golang.org/x/image/draw and CatmullRom interpolation. (Slow but high quality)
func ResizeImage(img image.Image, width, height uint) image.Image {
	if width == 0 || height == 0 {
		return img
	}
	rect := image.Rect(0, 0, int(width), int(height))
	resizedImg := image.NewRGBA(rect)
	draw.CatmullRom.Scale(resizedImg, rect, img, img.Bounds(), draw.Over, nil)
	return resizedImg
}

func GetImageSize(r io.Reader) (int, int, error) {
	img, _, err := image.DecodeConfig(r)
	if err != nil {
		return 0, 0, err
	}
	return img.Width, img.Height, nil
}

func CompressImageToJPEG(r io.Reader, maxSizeMB, maxEdgeLength uint, cacheKey string) ([]byte, error) {
	if cacheKey != "" {
		cachePath := filepath.Join(config.Cfg.Storage.CacheDir, "image", EscapeFileName(cacheKey))
		data, err := os.ReadFile(cachePath)
		if err == nil {
			return data, nil
		}
	}

	img, _, err := image.Decode(r)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()
	var newWidth, newHeight uint
	if width > int(maxEdgeLength) || height > int(maxEdgeLength) {
		if width > height {
			newWidth = maxEdgeLength
			newHeight = uint(float64(height) * (float64(maxEdgeLength) / float64(width)))
		} else {
			newHeight = maxEdgeLength
			newWidth = uint(float64(width) * (float64(maxEdgeLength) / float64(height)))
		}
		img = ResizeImage(img, newWidth, newHeight)
	}

	buf := imageBufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer func() {
		buf.Reset()
		imageBufferPool.Put(buf)
	}()
	maxSize := int(maxSizeMB * 1024 * 1024)
	if buf.Cap() < maxSize {
		buf.Grow(maxSize)
	}
	quality := 100
	for {
		if quality <= 0 {
			return nil, fmt.Errorf("cannot compress image to %d MB", maxSizeMB)
		}
		buf.Reset()
		err = jpeg.Encode(buf, img, &jpeg.Options{Quality: quality})
		if err != nil {
			return nil, fmt.Errorf("failed to encode image: %w", err)
		}
		if buf.Len() < maxSize {
			break
		}
		quality -= 5
	}

	result := bytes.Clone(buf.Bytes())

	if cacheKey != "" {
		go MkCache(filepath.Join(config.Cfg.Storage.CacheDir, "image", EscapeFileName(cacheKey)), result, time.Duration(config.Cfg.Storage.CacheTTL)*time.Second)
	}
	return result, nil
}

// 使用 ffmpeg 压缩图片
func CompressImageByFFmpeg(inputPath, outputPath string, maxEdgeLength uint, quality uint) error {
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
			vfKwArg = ffmpeg.KwArgs{"vf": fmt.Sprintf("scale=%d:-1", maxEdgeLength)}
		} else {
			vfKwArg = ffmpeg.KwArgs{"vf": fmt.Sprintf("scale=-1:%d", maxEdgeLength)}
		}
	}
	qualityArg := ffmpeg.KwArgs{"q": quality}
	if err := ffmpeg.Input(inputPath).Output(outputPath, vfKwArg, qualityArg).OverWriteOutput().Run(); err != nil {
		Logger.Errorf("failed to compress image: %s", err)
		return err
	}
	return nil
}
