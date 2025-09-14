package imgtool

import (
	"fmt"
	"image"

	"github.com/cshum/vipsgen/vips"
	"github.com/krau/ManyACG/types"
)

// var rgbaPool = sync.Pool{
// 	New: func() any {
// 		return &image.RGBA{}
// 	},
// }

// ResizeImage resizes an image to the specified width and height.
//
// It use golang.org/x/image/draw and CatmullRom interpolation. (Slow but high quality, and cost many memory)
// func ResizeImage(img image.Image, width, height uint) image.Image {
// 	if width == 0 || height == 0 {
// 		return img
// 	}

// 	rgba := rgbaPool.Get().(*image.RGBA)
// 	bounds := image.Rect(0, 0, int(width), int(height))

// 	if rgba.Bounds() != bounds {
// 		rgba = image.NewRGBA(bounds)
// 	}

// 	draw.CatmullRom.Scale(rgba, bounds, img, img.Bounds(), draw.Over, nil)

// 	return rgba
// }

// func GetImageSizeFromReader(r io.Reader) (int, int, error) {
// 	img, _, err := image.DecodeConfig(r)
// 	if err != nil {
// 		return 0, 0, err
// 	}
// 	return img.Width, img.Height, nil
// }

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

// var imageBufferPool = sync.Pool{
// 	New: func() any {
// 		return new(bytes.Buffer)
// 	},
// }
// via go native
// func CompressImageToJPEG(input []byte, maxEdgeLength, maxFileSize uint) ([]byte, error) {
// 	img, _, err := image.Decode(bytes.NewReader(input))
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to decode image: %w", err)
// 	}

// 	bounds := img.Bounds()
// 	width := bounds.Dx()
// 	height := bounds.Dy()
// 	var newWidth, newHeight uint
// 	if width > int(maxEdgeLength) || height > int(maxEdgeLength) {
// 		if width > height {
// 			newWidth = maxEdgeLength
// 			newHeight = uint(float64(height) * (float64(maxEdgeLength) / float64(width)))
// 		} else {
// 			newHeight = maxEdgeLength
// 			newWidth = uint(float64(width) * (float64(maxEdgeLength) / float64(height)))
// 		}
// 		img = ResizeImage(img, newWidth, newHeight)
// 	}

// 	buf := imageBufferPool.Get().(*bytes.Buffer)
// 	buf.Reset()
// 	defer func() {
// 		buf.Reset()
// 		imageBufferPool.Put(buf)
// 	}()
// 	if buf.Cap() < int(maxFileSize) {
// 		buf.Grow(int(maxFileSize) - buf.Cap())
// 	}
// 	quality := 100
// 	for {
// 		if quality <= 0 {
// 			return nil, fmt.Errorf("cannot compress image to %d MB", maxFileSize/1024/1024)
// 		}
// 		buf.Reset()
// 		err = jpeg.Encode(buf, img, &jpeg.Options{Quality: quality})
// 		if err != nil {
// 			return nil, fmt.Errorf("failed to encode image: %w", err)
// 		}
// 		if buf.Len() < int(maxFileSize) {
// 			break
// 		}
// 		quality -= 5
// 	}
// 	result := bytes.Clone(buf.Bytes())
// 	return result, nil
// }

// 使用 ffmpeg 压缩图片
// func CompressImageByFFmpeg(inputPath, outputPath string, maxEdgeLength int) error {
// 	file, err := os.Open(inputPath)
// 	if err != nil {
// 		return err
// 	}
// 	defer file.Close()
// 	img, _, err := image.DecodeConfig(file)
// 	if err != nil {
// 		return err
// 	}
// 	var vfKwArg ffmpeg.KwArgs
// 	if img.Width > int(maxEdgeLength) || img.Height > int(maxEdgeLength) {
// 		if img.Width > img.Height {
// 			vfKwArg = ffmpeg.KwArgs{"vf": fmt.Sprintf("scale=%d:-1:flags=lanczos", maxEdgeLength)}
// 		} else {
// 			vfKwArg = ffmpeg.KwArgs{"vf": fmt.Sprintf("scale=-1:%d:flags=lanczos", maxEdgeLength)}
// 		}
// 	}
// 	if err := ffmpeg.Input(inputPath).Output(outputPath, vfKwArg).OverWriteOutput().Run(); err != nil {
// 		return fmt.Errorf("failed to compress image: %w", err)
// 	}
// 	return nil
// }

func CompressImageByVIPS(inputPath, outputPath, format string, maxEdgeLength int) error {
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
		return fmt.Errorf("unsupported format: %s", format)
	}
	if err != nil {
		return fmt.Errorf("failed to save image to file: %w", err)
	}
	return nil
}

// by ffmpeg
// func CompressImageForTelegram(input []byte) ([]byte, error) {
// img, _, err := image.DecodeConfig(bytes.NewReader(input))
// if err != nil {
// 	return nil, fmt.Errorf("failed to decode image: %w", err)
// }
// inputLen := len(input)
// currentTotalSideLength := img.Width + img.Height
// if currentTotalSideLength <= types.TelegramMaxPhotoTotalSideLength && inputLen <= types.TelegramMaxPhotoFileSize {
// 	return input, nil
// }

// scaleFactor := float64(types.TelegramMaxPhotoTotalSideLength) / float64(currentTotalSideLength)
// newWidth := int(math.Floor(float64(img.Width) * scaleFactor))
// newHeight := int(math.Floor(float64(img.Height) * scaleFactor))
// if newWidth == 0 || newHeight == 0 {
// 	return nil, fmt.Errorf("failed to calculate new image size")
// }
// vfKwArg := ffmpeg.KwArgs{"vf": fmt.Sprintf("scale=%d:%d:flags=lanczos", newWidth, newHeight)}

// depth := 0
// for {
// 	if depth > 5 {
// 		return nil, fmt.Errorf("failed to compress image")
// 	}
// 	buf := bytes.NewBuffer(nil)
// 	err = ffmpeg.Input("pipe:").
// 		Output("pipe:", vfKwArg, ffmpeg.KwArgs{"format": "mjpeg"}, ffmpeg.KwArgs{"q:v": 2 + depth}).
// 		WithInput(bytes.NewReader(input)).
// 		WithOutput(buf).Run()
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to compress image: %w", err)
// 	}
// 	if buf.Len() > types.TelegramMaxPhotoFileSize {
// 		// Logger.Debugf("recompressing...;current: compressed image size: %.2f MB, input size: %.2f MB, depth: %d;", float64(buf.Len())/1024/1024, float64(inputLen)/1024/1024, depth)
// 		depth++
// 		continue
// 	}
// 	return buf.Bytes(), nil
// }
// }

func CompressImageForTelegram(input []byte) ([]byte, error) {
	img, err := vips.NewImageFromBuffer(input, vips.DefaultLoadOptions())
	if err != nil {
		return nil, fmt.Errorf("failed to create image from buffer: %w", err)
	}
	defer img.Close()
	width := img.Width()
	height := img.Height()

	var scale float64 = 1.0
	maxLength := types.RegularPhotoSideLength
	if width > height {
		if width > maxLength {
			scale = float64(maxLength) / float64(width)
		}
	} else {
		if height > maxLength {
			scale = float64(maxLength) / float64(height)
		}
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
