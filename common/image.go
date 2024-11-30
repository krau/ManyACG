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
	"math"
	"os"
	"strconv"

	"golang.org/x/image/draw"

	"sync"

	"github.com/corona10/goimagehash"
	"github.com/krau/ManyACG/types"
	ffmpeg "github.com/u2takey/ffmpeg-go"
)

func GetImagePhash(img image.Image) (string, error) {
	return getImagePhash(img)
}

func GetImagePhashFromReader(r io.Reader) (string, error) {
	img, _, err := image.Decode(r)
	if err != nil {
		return "", err
	}
	return getImagePhash(img)
}

func getImagePhash(img image.Image) (string, error) {
	hash, err := goimagehash.PerceptionHash(img)
	if err != nil {
		return "", err
	}
	return hash.ToString(), nil
}

// Deprecated: sometimes inaccurate
func GetImageBlurScore(img image.Image) (float64, error) {
	return getImageBlurScore(img)
}

// Deprecated: sometimes inaccurate
func GetImageBlurScoreFromReader(r io.Reader) (float64, error) {
	img, _, err := image.Decode(r)
	if err != nil {
		return 0, err
	}
	return getImageBlurScore(img)
}

func getImageBlurScore(img image.Image) (float64, error) {
	// TODO: use more accurate algorithm
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

var rgbaPool = sync.Pool{
	New: func() any {
		return &image.RGBA{}
	},
}

// ResizeImage resizes an image to the specified width and height.
//
// It use golang.org/x/image/draw and CatmullRom interpolation. (Slow but high quality, and cost many memory)
func ResizeImage(img image.Image, width, height uint) image.Image {
	if width == 0 || height == 0 {
		return img
	}

	rgba := rgbaPool.Get().(*image.RGBA)
	bounds := image.Rect(0, 0, int(width), int(height))

	if rgba.Bounds() != bounds {
		rgba = image.NewRGBA(bounds)
	}

	draw.CatmullRom.Scale(rgba, bounds, img, img.Bounds(), draw.Over, nil)

	return rgba
}

func GetImageSizeFromReader(r io.Reader) (int, int, error) {
	img, _, err := image.DecodeConfig(r)
	if err != nil {
		return 0, 0, err
	}
	return img.Width, img.Height, nil
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

var imageBufferPool = sync.Pool{
	New: func() interface{} {
		return new(bytes.Buffer)
	},
}

func CompressImageToJPEG(input []byte, maxEdgeLength, maxFileSize uint) ([]byte, error) {
	img, _, err := image.Decode(bytes.NewReader(input))
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
	if buf.Cap() < int(maxFileSize) {
		buf.Grow(int(maxFileSize) - buf.Cap())
	}
	quality := 100
	for {
		if quality <= 0 {
			return nil, fmt.Errorf("cannot compress image to %d MB", maxFileSize/1024/1024)
		}
		buf.Reset()
		err = jpeg.Encode(buf, img, &jpeg.Options{Quality: quality})
		if err != nil {
			return nil, fmt.Errorf("failed to encode image: %w", err)
		}
		if buf.Len() < int(maxFileSize) {
			break
		}
		quality -= 5
	}
	result := bytes.Clone(buf.Bytes())
	return result, nil
}

// 使用 ffmpeg 压缩图片
func CompressImageByFFmpeg(inputPath, outputPath string, maxEdgeLength int) error {
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
		Logger.Errorf("failed to compress image: %s", err)
		return err
	}
	return nil
}

func CompressImageForTelegramByFFmpegFromBytes(input []byte, maxDepth uint, extraFFmpegKwArgs ...ffmpeg.KwArgs) ([]byte, error) {
	if maxDepth == 0 {
		return nil, fmt.Errorf("max depth reached")
	}
	if extraFFmpegKwArgs == nil {
		extraFFmpegKwArgs = make([]ffmpeg.KwArgs, 0)
	}

	settedQV := false
	for _, kwArgs := range extraFFmpegKwArgs {
		if kwArgs.HasKey("vf") {
			return nil, fmt.Errorf("vf kwarg is not allowed in extraFFmpegKwArgs")
		}
		if kwArgs.HasKey("q:v") {
			settedQV = true
			if kwArgs.GetString("q:v") == "0" {
				delete(kwArgs, "q:v")
			}
		}
	}
	if !settedQV {
		extraFFmpegKwArgs = append(extraFFmpegKwArgs, ffmpeg.KwArgs{"q:v": 2})
	}

	img, _, err := image.DecodeConfig(bytes.NewReader(input))
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}
	inputLen := len(input)
	currentTotalSideLength := img.Width + img.Height
	if currentTotalSideLength <= types.TelegramMaxPhotoTotalSideLength && inputLen <= types.TelegramMaxPhotoFileSize {
		return input, nil
	}

	scaleFactor := float64(types.TelegramMaxPhotoTotalSideLength) / float64(currentTotalSideLength)
	newWidth := int(math.Round(float64(img.Width) * scaleFactor))
	newHeight := int(math.Round(float64(img.Height) * scaleFactor))
	if newWidth == 0 || newHeight == 0 {
		return nil, fmt.Errorf("failed to calculate new image size")
	}
	vfKwArg := ffmpeg.KwArgs{"vf": fmt.Sprintf("scale=%d:%d:flags=lanczos", newWidth, newHeight)}

	buf := bytes.NewBuffer(nil)

	err = ffmpeg.Input("pipe:").Output("pipe:", vfKwArg, ffmpeg.KwArgs{"format": "mjpeg"}, ffmpeg.MergeKwArgs(extraFFmpegKwArgs)).WithInput(bytes.NewReader(input)).WithOutput(buf).Run()

	if err != nil {
		return nil, fmt.Errorf("failed to compress image: %w", err)
	}
	if buf.Len() > inputLen {
		Logger.Warnf("compressed image file size %d is larger than original file size %d, drop quality parameter and retry", buf.Len(), inputLen)
		for i, kwArgs := range extraFFmpegKwArgs {
			if kwArgs.HasKey("q:v") {
				delete(extraFFmpegKwArgs[i], "q:v")
			}
		}
		extraFFmpegKwArgs = append(extraFFmpegKwArgs, ffmpeg.KwArgs{"q:v": 0})
		return CompressImageForTelegramByFFmpegFromBytes(input, maxDepth-1, ffmpeg.MergeKwArgs(extraFFmpegKwArgs))
	}
	if buf.Len() > types.TelegramMaxPhotoFileSize {
		return CompressImageForTelegramByFFmpegFromBytes(buf.Bytes(), maxDepth-1)
	}
	return buf.Bytes(), nil
}
