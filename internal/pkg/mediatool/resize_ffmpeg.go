package mediatool

import (
	"fmt"
	"image"
	"os"
	"strconv"

	"github.com/krau/ffmpeg-go"
)

// encodeTelegramPhotoByFFmpeg 将 plan 应用到图片并按指定格式编码.
// 内容叠加到白色画布上, 同时完成透明通道合成、缩放与居中加边.
func encodeTelegramPhotoByFFmpeg(inputPath, outputPath, format string, quality int, plan telegramPhotoPlan) error {
	background := ffmpeg.Input(
		fmt.Sprintf("color=c=white:s=%dx%d:r=1", plan.canvasWidth, plan.canvasHeight),
		ffmpeg.KwArgs{"f": "lavfi"},
	)
	content := ffmpeg.Input(inputPath)
	if plan.scale < 1 {
		content = content.Filter("scale", ffmpeg.Args{
			strconv.Itoa(plan.width), strconv.Itoa(plan.height),
		}, ffmpeg.KwArgs{"flags": "lanczos"})
	}
	content = content.Filter("format", ffmpeg.Args{"rgba"})
	left := (plan.canvasWidth - plan.width) / 2
	top := (plan.canvasHeight - plan.height) / 2
	composed := background.Overlay(content, "", ffmpeg.KwArgs{
		"x":      strconv.Itoa(left),
		"y":      strconv.Itoa(top),
		"format": "rgb",
	}).Filter("format", ffmpeg.Args{"rgb24"})
	kwargs := ffmpeg.KwArgs{"frames:v": "1"}
	switch format {
	case "png":
		kwargs["c:v"] = "png"
	case "jpeg":
		kwargs["q:v"] = strconv.Itoa(ffmpegJPEGQuality(quality))
	default:
		return fmt.Errorf("unsupported image format: %s", format)
	}
	if err := composed.Output(outputPath, kwargs).OverWriteOutput().Run(); err != nil {
		return fmt.Errorf("failed to compress image: %w", err)
	}
	return nil
}

// ffmpegJPEGQuality 将 JPEG 质量(1-100)映射为 ffmpeg mjpeg 的 -q:v 值.
func ffmpegJPEGQuality(quality int) int {
	switch {
	case quality >= 95:
		return 2
	case quality >= 90:
		return 3
	case quality >= 85:
		return 4
	default:
		return 5
	}
}

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
	if maxEdgeLength > 0 {
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
	} else {
		if err := ffmpeg.Input(inputPath).Output(outputPath).OverWriteOutput().Run(); err != nil {
			return fmt.Errorf("failed to compress image: %w", err)
		}
	}
	return nil
}
