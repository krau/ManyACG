package mediatool

import (
	"fmt"
	"image"
	"os"
	"strings"

	"github.com/krau/ffmpeg-go"
)

// 使用 ffmpeg 将图片压缩为可发送的 Telegram photo
func compressImageForTelegramByFFmpeg(inputPath, outputPath string) error {
	file, err := os.Open(inputPath)
	if err != nil {
		return err
	}
	defer file.Close()
	img, _, err := image.DecodeConfig(file)
	if err != nil {
		return err
	}
	plan := planTelegramPhoto(img.Width, img.Height, TelegramMaxPhotoSideLength)
	var vfKwArg ffmpeg.KwArgs
	if filter := telegramPhotoFilter(plan); filter != "" {
		vfKwArg = ffmpeg.KwArgs{"vf": filter}
	}
	if err := ffmpeg.Input(inputPath).Output(outputPath, vfKwArg).OverWriteOutput().Run(); err != nil {
		return fmt.Errorf("failed to compress image: %w", err)
	}
	return nil
}

// telegramPhotoFilter 返回缩放并居中加边的 ffmpeg 滤镜链, 无需变换时返回空字符串.
func telegramPhotoFilter(plan telegramPhotoPlan) string {
	filters := make([]string, 0, 2)
	if plan.scale < 1 {
		filters = append(filters, fmt.Sprintf("scale=%d:%d:flags=lanczos", plan.width, plan.height))
	}
	if plan.width != plan.canvasWidth || plan.height != plan.canvasHeight {
		left := (plan.canvasWidth - plan.width) / 2
		top := (plan.canvasHeight - plan.height) / 2
		filters = append(filters, fmt.Sprintf("pad=%d:%d:%d:%d:color=white", plan.canvasWidth, plan.canvasHeight, left, top))
	}
	return strings.Join(filters, ",")
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
