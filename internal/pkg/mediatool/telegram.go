package mediatool

import (
	"bytes"
	"fmt"
	"image"
	"io"
	"os"
	"path/filepath"

	"github.com/krau/ManyACG/pkg/osutil"
	"github.com/krau/ManyACG/pkg/strutil"
)

var pngSignature = []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n'}

// telegramPhotoPlan 描述一张图片发送为 Telegram photo 所需的变换.
type telegramPhotoPlan struct {
	scale        float64 // 等比缩放系数
	width        int     // 缩放后的图片宽高
	height       int
	canvasWidth  int // 最终画布宽高, 图片居中放置
	canvasHeight int
}

// needsTransform 表示图片需要缩放或加边后才能发送.
func (p telegramPhotoPlan) needsTransform() bool {
	return p.scale < 1 || p.width != p.canvasWidth || p.height != p.canvasHeight
}

// planTelegramPhoto 计算 width x height 的图片发送为 Telegram photo 所需的变换,
// 限制见 https://core.telegram.org/bots/api#sendphoto :
// 宽高之和不大于 TelegramMaxPhotoTotalSideLength, 宽高比不大于 TelegramMaxPhotoAspectRatio.
//
// 长边限制在 TelegramMaxPhotoSideLength (服务端保存的最大档位), 宽高和因此必然满足限制.
// 宽高比超限的图片无法通过缩放修复, 只能加边(letterbox), 且不做放大.
func planTelegramPhoto(width, height int) telegramPhotoPlan {
	plan := telegramPhotoPlan{
		scale:        1,
		width:        width,
		height:       height,
		canvasWidth:  width,
		canvasHeight: height,
	}
	if width <= 0 || height <= 0 {
		return plan
	}
	if long := max(width, height); long > TelegramMaxPhotoSideLength {
		plan.scale = float64(TelegramMaxPhotoSideLength) / float64(long)
	}
	plan.width = max(int(float64(width)*plan.scale), 1)
	plan.height = max(int(float64(height)*plan.scale), 1)
	plan.canvasWidth, plan.canvasHeight = plan.width, plan.height
	long, short := max(plan.width, plan.height), min(plan.width, plan.height)
	if long > TelegramMaxPhotoAspectRatio*short {
		if plan.width >= plan.height {
			plan.canvasHeight = paddedShortSide(plan.width)
		} else {
			plan.canvasWidth = paddedShortSide(plan.height)
		}
	}
	return plan
}

// TelegramPhotoSize 返回 width x height 的图片作为 Telegram photo 发送时的最终尺寸,
// 供外部(如图片代理)按此尺寸缩放加边.
func TelegramPhotoSize(width, height int) (int, int) {
	plan := planTelegramPhoto(width, height)
	return plan.canvasWidth, plan.canvasHeight
}

// paddedShortSide 返回使宽高比严格小于 TelegramMaxPhotoAspectRatio 的最短边.
func paddedShortSide(long int) int {
	short := ceilDiv(long, TelegramMaxPhotoAspectRatio)
	if short*TelegramMaxPhotoAspectRatio == long {
		short++
	}
	return short
}

func ceilDiv(a, b int) int {
	return (a + b - 1) / b
}

// telegramPhotoSource 是源文件中与 Telegram photo 限制相关的信息.
type telegramPhotoSource struct {
	format   string // 仅识别 Telegram 接受的 jpeg / png
	hasAlpha bool
}

// inspectTelegramPhotoSource 通过文件头识别格式与透明通道.
func inspectTelegramPhotoSource(filePath string) telegramPhotoSource {
	f, err := os.Open(filePath)
	if err != nil {
		return telegramPhotoSource{}
	}
	defer f.Close()
	header := make([]byte, 32)
	n, _ := io.ReadFull(f, header)
	header = header[:n]
	switch {
	case len(header) >= 3 && header[0] == 0xff && header[1] == 0xd8 && header[2] == 0xff:
		return telegramPhotoSource{format: "jpeg"}
	case len(header) >= 26 && bytes.Equal(header[:8], pngSignature):
		colorType := header[25]
		return telegramPhotoSource{format: "png", hasAlpha: colorType == 4 || colorType == 6}
	}
	return telegramPhotoSource{}
}

// prepareTelegramPhoto 生成可发送的 Telegram photo 并返回输出路径.
//
// 保证输出满足 Telegram 的大小/尺寸/比例/格式限制, 并尽量保留画质:
// 无需变换的 JPEG/PNG 原样保留, 否则优先编码为无损 PNG, 超出体积限制时退回高质量 JPEG.
func prepareTelegramPhoto(inputPath, outputDir string) (string, error) {
	inputInfo, err := os.Stat(inputPath)
	if err != nil {
		return "", err
	}
	width, height, err := imageSize(inputPath)
	if err != nil {
		return "", fmt.Errorf("failed to get image size: %w", err)
	}
	plan := planTelegramPhoto(width, height)
	source := inspectTelegramPhotoSource(inputPath)
	if !plan.needsTransform() && !source.hasAlpha && source.format != "" &&
		inputInfo.Size() <= int64(TelegramMaxPhotoFileSize) {
		outputPath := telegramPhotoOutputPath(inputPath, outputDir, source.format)
		return outputPath, copyFile(inputPath, outputPath)
	}
	attempts := []struct {
		format  string
		quality int
	}{
		{"png", 0},
		{"jpeg", 95},
		{"jpeg", 90},
		{"jpeg", 85},
	}
	var lastErr error
	for _, attempt := range attempts {
		outputPath := telegramPhotoOutputPath(inputPath, outputDir, attempt.format)
		err := encodeTelegramPhoto(inputPath, outputPath, attempt.format, attempt.quality, plan)
		if err != nil {
			lastErr = err
			os.Remove(outputPath)
			continue
		}
		info, err := os.Stat(outputPath)
		if err != nil {
			return "", err
		}
		if info.Size() > int64(TelegramMaxPhotoFileSize) {
			lastErr = fmt.Errorf("%s output exceeds %d bytes", attempt.format, TelegramMaxPhotoFileSize)
			os.Remove(outputPath)
			continue
		}
		if err := verifyTelegramPhoto(outputPath, plan); err != nil {
			lastErr = err
			os.Remove(outputPath)
			continue
		}
		return outputPath, nil
	}
	return "", fmt.Errorf("failed to compress image for telegram: %w", lastErr)
}

// verifyTelegramPhoto 在发送前校验输出文件, 不满足限制的结果不会被发送出去.
func verifyTelegramPhoto(filePath string, plan telegramPhotoPlan) error {
	info, err := os.Stat(filePath)
	if err != nil {
		return err
	}
	if info.Size() > int64(TelegramMaxPhotoFileSize) {
		return fmt.Errorf("size %d exceeds %d", info.Size(), TelegramMaxPhotoFileSize)
	}
	if source := inspectTelegramPhotoSource(filePath); source.format == "" {
		return fmt.Errorf("unsupported telegram photo format")
	}
	width, height, err := imageSize(filePath)
	if err != nil {
		return fmt.Errorf("failed to get image size: %w", err)
	}
	if width != plan.canvasWidth || height != plan.canvasHeight {
		return fmt.Errorf("dimensions %dx%d, want %dx%d", width, height, plan.canvasWidth, plan.canvasHeight)
	}
	if width+height > TelegramMaxPhotoTotalSideLength {
		return fmt.Errorf("dimensions %dx%d exceed %d in total", width, height, TelegramMaxPhotoTotalSideLength)
	}
	if max(width, height) > TelegramMaxPhotoAspectRatio*min(width, height) {
		return fmt.Errorf("ratio %dx%d exceeds %d", width, height, TelegramMaxPhotoAspectRatio)
	}
	return nil
}

// encodeTelegramPhoto 将 plan 应用到图片并按指定格式编码.
func encodeTelegramPhoto(inputPath, outputPath, format string, quality int, plan telegramPhotoPlan) error {
	switch {
	case vipsJPEGAvailable():
		return encodeTelegramPhotoByVIPS(inputPath, outputPath, format, quality, plan)
	case ffmpegAvailable:
		return encodeTelegramPhotoByFFmpeg(inputPath, outputPath, format, quality, plan)
	default:
		return encodeTelegramPhotoNative(inputPath, outputPath, format, quality, plan)
	}
}

// telegramPhotoOutputPath 返回与输出格式匹配的临时路径.
func telegramPhotoOutputPath(inputPath, outputDir, format string) string {
	ext := "png"
	if format == "jpeg" {
		ext = "jpg"
	}
	outputPath := filepath.Join(outputDir, fmt.Sprintf("tg_%s.%s", strutil.MD5Hash(inputPath), ext))
	return osutil.TempSiblingPath(outputPath)
}

// imageSize 返回图片宽高, 优先使用标准库解码器, 无法识别时回退到 libvips.
func imageSize(filePath string) (int, int, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return 0, 0, err
	}
	cfg, _, decodeErr := image.DecodeConfig(f)
	f.Close()
	if decodeErr == nil {
		return cfg.Width, cfg.Height, nil
	}
	if width, height, err := vipsImageSize(filePath); err == nil {
		return width, height, nil
	}
	return 0, 0, decodeErr
}

func copyFile(inputPath, outputPath string) error {
	if err := os.MkdirAll(filepath.Dir(outputPath), os.ModePerm); err != nil {
		return err
	}
	in, err := os.Open(inputPath)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		return err
	}
	return out.Close()
}
