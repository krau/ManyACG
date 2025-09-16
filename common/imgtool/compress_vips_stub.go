//go:build !linux || !amd64

package imgtool

import "fmt"

func compressImageVIPS(inputPath, outputPath, format string, maxEdgeLength int) error {
	return fmt.Errorf("vips compression is only supported on linux/amd64")
}

func compressImageForTelegramByVIPS(input []byte) ([]byte, error) {
	return nil, fmt.Errorf("vips compression is only supported on linux/amd64")
}
