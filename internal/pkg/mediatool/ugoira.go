package mediatool

import (
	"archive/zip"
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/krau/ManyACG/internal/shared"
	"github.com/krau/ManyACG/pkg/log"
	"github.com/krau/ffmpeg-go"
)

// UgoiraZipToMp4 从 ugoira 的 zip 解压并转换为 mp4
//
// 返回生成的 mp4 路径（可能与 outputPath 不同，会自动添加 .mp4 后缀）
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
		// duration 单位为（小数）
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

	// ffmpeg -f concat -safe 0 -i ffconcat.txt -vsync vfr -pix_fmt yuv420p -c:v libx264 output.mp4
	in := ffmpeg.Input(listPath, ffmpeg.KwArgs{
		"f":    "concat",
		"safe": "0",
	})

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
