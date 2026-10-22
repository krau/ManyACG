//go:build !vips

package mediatool

import "fmt"

func compressImageVIPS(inputPath, outputPath, format string, maxEdgeLength int) error {
	return fmt.Errorf("vips compression is only supported on linux/amd64")
}

func encodeTelegramPhotoByVIPS(inputPath, outputPath, format string, quality int, plan telegramPhotoPlan) error {
	return fmt.Errorf("vips compression is only supported on linux/amd64")
}

func vipsImageSize(filePath string) (int, int, error) {
	return 0, 0, fmt.Errorf("vips is only supported on linux/amd64")
}
