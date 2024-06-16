package common

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	_ "image/gif"
	"image/jpeg"
	_ "image/png"
	"strconv"

	"golang.org/x/image/draw"

	"github.com/corona10/goimagehash"
)

func GetPhash(b []byte) (string, error) {
	r := bytes.NewReader(b)
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

func GetBlurScore(b []byte) (float64, error) {
	r := bytes.NewReader(b)
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

func ResizeImage(img image.Image, width, height uint) image.Image {
	rect := image.Rect(0, 0, int(width), int(height))
	resizedImg := image.NewRGBA(rect)
	draw.CatmullRom.Scale(resizedImg, rect, img, img.Bounds(), draw.Over, nil)
	return resizedImg
}

func CompressImage(input []byte, maxSizeMB, maxEdgeLength uint) ([]byte, error) {
	img, _, err := image.Decode(bytes.NewReader(input))
	if err != nil {
		return nil, err
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
	quality := 100
	var buf bytes.Buffer
	err = jpeg.Encode(&buf, img, &jpeg.Options{Quality: quality})
	if err != nil {
		return nil, err
	}
	if buf.Len() < int(maxSizeMB*1024*1024) {
		return buf.Bytes(), nil
	}
	for {
		if quality <= 0 {
			return nil, fmt.Errorf("cannot compress image to %d MB", maxSizeMB)
		}
		quality -= 5
		buf.Reset()
		err = jpeg.Encode(&buf, img, &jpeg.Options{Quality: quality})
		if err != nil {
			return nil, err
		}
		if buf.Len() < int(maxSizeMB*1024*1024) {
			break
		}
	}
	return buf.Bytes(), nil
}

func GetImageSize(b []byte) (int, int, error) {
	r := bytes.NewReader(b)
	img, _, err := image.DecodeConfig(r)
	if err != nil {
		return 0, 0, err
	}
	return img.Width, img.Height, nil
}
