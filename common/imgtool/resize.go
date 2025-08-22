package imgtool

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"math"
	"os"
	"sync"

	"github.com/krau/ManyACG/types"
	ffmpeg "github.com/u2takey/ffmpeg-go"
	"golang.org/x/image/draw"
)

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
		return fmt.Errorf("failed to compress image: %w", err)
	}
	return nil
}

func CompressImageForTelegramByFFmpegFromBytes(input []byte) ([]byte, error) {
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
	newWidth := int(math.Floor(float64(img.Width) * scaleFactor))
	newHeight := int(math.Floor(float64(img.Height) * scaleFactor))
	if newWidth == 0 || newHeight == 0 {
		return nil, fmt.Errorf("failed to calculate new image size")
	}
	vfKwArg := ffmpeg.KwArgs{"vf": fmt.Sprintf("scale=%d:%d:flags=lanczos", newWidth, newHeight)}

	depth := 0
	for {
		if depth > 5 {
			return nil, fmt.Errorf("failed to compress image")
		}
		buf := bytes.NewBuffer(nil)
		err = ffmpeg.Input("pipe:").
			Output("pipe:", vfKwArg, ffmpeg.KwArgs{"format": "mjpeg"}, ffmpeg.KwArgs{"q:v": 2 + depth}).
			WithInput(bytes.NewReader(input)).
			WithOutput(buf).Run()
		if err != nil {
			return nil, fmt.Errorf("failed to compress image: %w", err)
		}
		if buf.Len() > types.TelegramMaxPhotoFileSize {
			// Logger.Debugf("recompressing...;current: compressed image size: %.2f MB, input size: %.2f MB, depth: %d;", float64(buf.Len())/1024/1024, float64(inputLen)/1024/1024, depth)
			depth++
			continue
		}
		return buf.Bytes(), nil
	}
}
