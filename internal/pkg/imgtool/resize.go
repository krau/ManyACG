package imgtool

import (
	"archive/zip"
	"bufio"
	"fmt"
	"image"
	"io"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/gen2brain/avif"
	"github.com/krau/ManyACG/internal/infra/config/runtimecfg"
	"github.com/krau/ManyACG/internal/shared"
	"github.com/krau/ManyACG/pkg/log"
	"github.com/krau/ManyACG/pkg/strutil"
	ffmpeg "github.com/u2takey/ffmpeg-go"
)

var (
	ffmpegAvailable bool
	vipsFormat      map[string]struct{}
	nativeFormat    = map[string]struct{}{"jpeg": {}, "jpg": {}, "png": {}, "webp": {}, "avif": {}}
)

func init() {
	avif.InitEncoder()
	avif.InitDecoder()
	switch runtime.GOOS {
	case "windows":
		_, err := exec.LookPath("ffmpeg.exe")
		if err == nil {
			ffmpegAvailable = true
		}
	default:
		_, err := exec.LookPath("ffmpeg")
		if err == nil {
			ffmpegAvailable = true
		}
	}
}

func FFmpegAvailable() bool {
	return ffmpegAvailable
}

func GetSize(img image.Image) (int, int, error) {
	if img == nil {
		return 0, 0, fmt.Errorf("nil image")
	}
	bounds := img.Bounds()
	if bounds.Empty() {
		return 0, 0, fmt.Errorf("empty image")
	}
	return bounds.Dx(), bounds.Dy(), nil
}

func GetSizeFromReader(r io.Reader) (int, int, error) {
	img, _, err := image.Decode(r)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to decode image: %w", err)
	}
	return GetSize(img)
}

func Compress(inputPath, outputPath, format string, maxEdgeLength int) error {
	if err := os.MkdirAll(filepath.Dir(outputPath), os.ModePerm); err != nil {
		return err
	}
	if _, ok := vipsFormat[format]; ok {
		log.Debug("compressing image", "method", "vips", "input", inputPath, "output", outputPath, "format", format)
		err := compressImageVIPS(inputPath, outputPath, format, maxEdgeLength)
		if err != nil {
			return fmt.Errorf("failed to compress image with vips: %w", err)
		}
		return nil
	}
	if ffmpegAvailable {
		log.Debug("compressing image", "method", "ffmpeg", "input", inputPath, "output", outputPath, "format", format)
		err := compressImageByFFmpeg(inputPath, outputPath, maxEdgeLength)
		if err != nil {
			return fmt.Errorf("failed to compress image with ffmpeg: %w", err)
		}
		return nil
	}
	if _, ok := nativeFormat[format]; ok {
		log.Debug("compressing image", "method", "native", "input", inputPath, "output", outputPath, "format", format)
		err := compressImageNative(inputPath, outputPath, format, maxEdgeLength)
		if err != nil {
			return fmt.Errorf("failed to compress image with native: %w", err)
		}
		return nil
	}
	return fmt.Errorf("unsupported image format: %s", format)
}

func CompressForTelegram(input []byte) ([]byte, error) {
	if _, ok := vipsFormat["jpeg"]; ok {
		return compressImageForTelegramByVIPS(input)
	}
	tmpFile, err := os.CreateTemp(runtimecfg.Get().Storage.CacheDir, "imgtool_*.png")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	distFile, err := os.CreateTemp(runtimecfg.Get().Storage.CacheDir, "imgtool_*.jpg")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(distFile.Name())
	defer distFile.Close()

	err = os.WriteFile(tmpFile.Name(), input, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to write temp file: %w", err)
	}
	if ffmpegAvailable {
		err = compressImageByFFmpeg(tmpFile.Name(), distFile.Name(), TelegramMaxPhotoSideLength)
		if err != nil {
			return nil, fmt.Errorf("failed to compress image by ffmpeg: %w", err)
		}
		result, err := os.ReadFile(distFile.Name())
		if err != nil {
			return nil, fmt.Errorf("failed to read temp file: %w", err)
		}
		return result, nil
	}
	err = compressImageNative(tmpFile.Name(), distFile.Name(), "jpeg", TelegramMaxPhotoSideLength)
	if err != nil {
		return nil, fmt.Errorf("failed to compress image natively: %w", err)
	}
	result, err := os.ReadFile(distFile.Name())
	if err != nil {
		return nil, fmt.Errorf("failed to read temp file: %w", err)
	}
	return result, nil
}

// TempFile is a temporary file that will be deleted when closed.
type TempFile struct {
	*os.File
}

func (t *TempFile) Close() error {
	err := t.File.Close()
	if err != nil {
		return err
	}
	return os.Remove(t.File.Name())
}

func CompressForTelegramFromFile(filePath string) (*TempFile, error) {
	outputPath := filepath.Join(runtimecfg.Get().Storage.CacheDir, "compress", fmt.Sprintf("tg_%s_%d.jpg", strutil.MD5Hash(filePath), rand.Int()))
	if _, ok := vipsFormat["jpeg"]; ok {
		err := compressImageForTelegramByVIPSFromFile(filePath, outputPath)
		if err != nil {
			return nil, err
		}
		f, err := os.Open(outputPath)
		if err != nil {
			return nil, err
		}
		return &TempFile{f}, nil
	}
	if ffmpegAvailable {
		err := compressImageByFFmpeg(filePath, outputPath, TelegramMaxPhotoSideLength)
		if err != nil {
			return nil, err
		}
		f, err := os.Open(outputPath)
		if err != nil {
			return nil, err
		}
		return &TempFile{f}, nil
	}
	err := compressImageNative(filePath, outputPath, "jpeg", TelegramMaxPhotoSideLength)
	if err != nil {
		return nil, err
	}
	f, err := os.Open(outputPath)
	if err != nil {
		return nil, err
	}
	return &TempFile{f}, nil
}

// UgoiraZipToMp4 从 ugoira 的 zip 解压并转换为 mp4
// zipPath: zip 文件路径
// frames: 按显示顺序的帧信息（File = zip 内文件名）
// outputPath: 目标 mp4 路径（会覆盖同名文件）
// 返回生成的 mp4 路径（可能与 outputPath 不同，因会自动添加 .mp4 后缀）
func UgoiraZipToMp4(zipPath string, frames []shared.UgoiraFrame, outputPath string) (string, error) {
	if !ffmpegAvailable {
		return "", fmt.Errorf("ffmpeg is not available")
	}
	tmpDir, err := os.MkdirTemp(filepath.Dir(outputPath), "manyacg-ugoira-*")
	if err != nil {
		return "", fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	extractedPaths := make([]string, 0, len(frames))
	// extract frames
	err = func() error {
		zr, err := zip.OpenReader(zipPath)
		if err != nil {
			return fmt.Errorf("open zip: %w", err)
		}
		defer zr.Close()

		zipMap := make(map[string]*zip.File)
		for _, f := range zr.File {
			zipMap[filepath.Base(f.Name)] = f
		}
		for i, fr := range frames {
			err := func() error {
				bname := filepath.Base(fr.File)
				zf, ok := zipMap[bname]
				if !ok {
					// 尝试按原名直接匹配（有时 frames.File 已带相对路径）
					zf = nil
					for _, f := range zr.File {
						if f.Name == fr.File || filepath.Base(f.Name) == fr.File {
							zf = f
							break
						}
					}
					if zf == nil {
						return fmt.Errorf("frame %d: file %q not found in zip", i, fr.File)
					}
				}
				// 解压
				rc, err := zf.Open()
				if err != nil {
					return fmt.Errorf("open zip entry %s: %w", zf.Name, err)
				}
				defer rc.Close()

				outPath := filepath.Join(tmpDir, bname)
				outFile, err := os.Create(outPath)
				if err != nil {
					return fmt.Errorf("create extracted file %s: %w", outPath, err)
				}
				_, err = io.Copy(outFile, rc)
				outFile.Close()
				if err != nil {
					return fmt.Errorf("write extracted file %s: %w", outPath, err)
				}

				extractedPaths = append(extractedPaths, outPath)
				return nil
			}()
			if err != nil {
				return err
			}
		}
		return nil
	}()
	if err != nil {
		return "", fmt.Errorf("extract frames: %w", err)
	}
	if len(extractedPaths) == 0 {
		return "", fmt.Errorf("no frames extracted")
	}
	// 生成 concat list 文件
	listPath := filepath.Join(tmpDir, "ffconcat.txt")
	listF, err := os.Create(listPath)
	if err != nil {
		return "", fmt.Errorf("create ffmpeg list file: %w", err)
	}
	bw := bufio.NewWriter(listF)
	for i, fr := range frames {
		if i >= len(extractedPaths) {
			log.Warn("ugoira frames length mismatch", "expected", len(frames), "got", len(extractedPaths))
			break
		}
		// duration 单位为秒（小数）
		delaySec := float64(fr.Delay) / 1000
		bw.WriteString(fmt.Sprintf("file '%s'\n", escapePathForConcat(extractedPaths[i])))
		bw.WriteString(fmt.Sprintf("duration %.6f\n", delaySec))
	}
	if len(extractedPaths) == 0 {
		listF.Close()
		return "", fmt.Errorf("no frames to encode")
	}
	// 重复一次最后的 file 行（concat demuxer 要求）
	// _, _ = bw.WriteString(fmt.Sprintf("file '%s'\n", escapePathForConcat(extractedPaths[len(extractedPaths)-1])))
	bw.Flush()
	listF.Close()

	// 5) 调用 ffmpeg-go (concat demuxer)
	// ffmpeg -f concat -safe 0 -i ffconcat.txt -vsync vfr -pix_fmt yuv420p -c:v libx264 output.mp4
	in := ffmpeg.Input(listPath, ffmpeg.KwArgs{
		"f":    "concat",
		"safe": "0",
	})

	// 检查 outputPath 是否以 .mp4 结尾
	ffoutPath := outputPath
	if strings.ToLower(filepath.Ext(outputPath)) != ".mp4" {
		ffoutPath += ".mp4"
	}

	// 调整为偶数边长, mp4 编码要求
	filtered := in.Filter("pad", ffmpeg.Args{
		"ceil(iw/2)*2", // width
		"ceil(ih/2)*2", // height
		"(ow-iw)/2",    // x offset (center)
		"(oh-ih)/2",    // y offset (center)
		"black",        // padding color
	})

	out := filtered.Output(ffoutPath, ffmpeg.KwArgs{
		"fps_mode": "vfr",
		"crf":      "23",
		"c:v":      "libx264",
		"pix_fmt":  "yuv420p",
	})
	// 覆盖输出
	if err := out.OverWriteOutput().ErrorToStdOut().Run(); err != nil {
		return "", fmt.Errorf("ffmpeg run error: %w", err)
	}
	return ffoutPath, nil
}

func escapePathForConcat(p string) string {
	abs, _ := filepath.Abs(p)
	// ffmpeg concat 需要单引号包裹，转义单引号
	return strings.ReplaceAll(abs, "'", "'\\''")
}
