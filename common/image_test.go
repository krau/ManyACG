package common

import (
	"image"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"path/filepath"
	"testing"
)

func loadImages(dir string) ([]image.Image, error) {
	var images []image.Image

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()

			img, _, err := image.Decode(file)
			if err != nil {
				return nil
			}

			images = append(images, img)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return images, nil
}

func BenchmarkGetImageBlurScore_RealImages(b *testing.B) {
	dir := filepath.Join("cache")

	images, err := loadImages(dir)
	if err != nil {
		b.Fatalf("Failed to load images from directory: %v", err)
	}
	if len(images) == 0 {
		b.Fatalf("No images found in directory: %s", dir)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		img := images[i%len(images)]
		_, err := getImageBlurScore(img)
		if err != nil {
			b.Fatalf("Error occurred: %v", err)
		}
	}
}
