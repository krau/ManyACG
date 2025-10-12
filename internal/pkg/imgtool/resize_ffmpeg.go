package imgtool

import (
	"fmt"
	"image"
	"os"

	"github.com/krau/ffmpeg-go"
)

// 使用 ffmpeg 压缩图片
func compressImageByFFmpeg(inputPath, outputPath string, maxEdgeLength int) error {
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
