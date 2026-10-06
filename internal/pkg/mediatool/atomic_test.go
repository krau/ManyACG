package mediatool

import (
	"archive/zip"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/krau/ManyACG/internal/shared"
	"github.com/krau/ManyACG/pkg/osutil"
)

func writeTestPNG(t *testing.T, path string, width, height int) {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := range height {
		for x := range width {
			img.Set(x, y, color.RGBA{R: uint8(x % 256), G: uint8(y % 256), B: 128, A: 255})
		}
	}
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("create png: %v", err)
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		t.Fatalf("encode png: %v", err)
	}
}

func writeTestUgoiraZip(t *testing.T, path string, frameSize int) []shared.UgoiraFrame {
	t.Helper()
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("create zip: %v", err)
	}
	zw := zip.NewWriter(f)
	frames := make([]shared.UgoiraFrame, 0, 3)
	for i := range 3 {
		name := fmt.Sprintf("frame_%03d.png", i)
		w, err := zw.Create(name)
		if err != nil {
			t.Fatalf("create zip entry: %v", err)
		}
		img := image.NewRGBA(image.Rect(0, 0, frameSize, frameSize))
		for y := range frameSize {
			for x := range frameSize {
				img.Set(x, y, color.RGBA{R: uint8(x * (i + 1)), G: uint8(y), B: uint8(i * 40), A: 255})
			}
		}
		if err := png.Encode(w, img); err != nil {
			t.Fatalf("encode frame: %v", err)
		}
		frames = append(frames, shared.UgoiraFrame{File: name, Delay: 100})
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("close zip: %v", err)
	}
	if err := f.Close(); err != nil {
		t.Fatalf("close zip file: %v", err)
	}
	return frames
}

// 压缩输出是临时产物, 并发压缩同一目标时不应写入同一路径
func TestCompressImgUsesUniqueOutput(t *testing.T) {
	dir := t.TempDir()
	inputPath := filepath.Join(dir, "input.png")
	writeTestPNG(t, inputPath, 400, 300)
	basePath := filepath.Join(dir, "output.png")

	const workers = 4
	paths := make([]string, workers)
	errs := make([]error, workers)
	var wg sync.WaitGroup
	for i := range workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			paths[i], errs[i] = CompressImg(inputPath, basePath, "png", 0)
		}()
	}
	wg.Wait()

	seen := make(map[string]struct{}, workers)
	for i := range workers {
		if errs[i] != nil {
			t.Fatalf("CompressImg: %v", errs[i])
		}
		if paths[i] == basePath {
			t.Fatalf("output is written to the shared path %s", paths[i])
		}
		if _, ok := seen[paths[i]]; ok {
			t.Fatalf("duplicated output path %s", paths[i])
		}
		seen[paths[i]] = struct{}{}
		if filepath.Ext(paths[i]) != ".png" {
			t.Errorf("unexpected output extension: %s", paths[i])
		}
		f, err := os.Open(paths[i])
		if err != nil {
			t.Fatalf("open output: %v", err)
		}
		cfg, format, err := image.DecodeConfig(f)
		f.Close()
		if err != nil {
			t.Fatalf("output %s is not a valid image: %v", paths[i], err)
		}
		if format != "png" || cfg.Width != 400 || cfg.Height != 300 {
			t.Errorf("unexpected output %s: format=%s size=%dx%d", paths[i], format, cfg.Width, cfg.Height)
		}
		if osutil.IsTempFile(paths[i]) {
			t.Errorf("output path looks like a temp file: %s", paths[i])
		}
	}
	assertNoTempFiles(t, dir)
}

func TestCompressImgFailureWritesNothing(t *testing.T) {
	for _, tt := range []struct {
		name   string
		base   string
		format string
		write  func(dir string) string
	}{
		{
			name: "unsupported format", base: "output.unknown", format: "unknown-format",
			write: func(dir string) string {
				p := filepath.Join(dir, "input.png")
				writeTestPNG(t, p, 32, 32)
				return p
			},
		},
		{
			name: "unreadable input", base: "output.png", format: "png",
			write: func(dir string) string {
				p := filepath.Join(dir, "broken.png")
				if err := os.WriteFile(p, []byte("not an image"), 0o644); err != nil {
					t.Fatal(err)
				}
				return p
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			inputPath := tt.write(dir)
			if _, err := CompressImg(inputPath, filepath.Join(dir, tt.base), tt.format, 0); err == nil {
				t.Fatal("expected an error")
			}
			assertNoTempFiles(t, dir)
			assertOnlyFiles(t, dir, filepath.Base(inputPath))
		})
	}
}

// 转换失败时不应留下产物或临时文件
func TestUgoiraZipToMp4FailureWritesNothing(t *testing.T) {
	if !FFmpegAvailable() {
		t.Skip("ffmpeg is not available")
	}
	dir := t.TempDir()
	zipPath := filepath.Join(dir, "frames.zip")
	f, err := os.Create(zipPath)
	if err != nil {
		t.Fatal(err)
	}
	zw := zip.NewWriter(f)
	w, err := zw.Create("frame_000.png")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := w.Write([]byte("not an image")); err != nil {
		t.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	frames := []shared.UgoiraFrame{{File: "frame_000.png", Delay: 100}}
	if _, err := UgoiraZipToMp4(zipPath, frames, filepath.Join(dir, "out.mp4")); err == nil {
		t.Fatal("expected an error for an undecodable frame")
	}
	assertNoTempFiles(t, dir)
	assertOnlyFiles(t, dir, "frames.zip")
}

// 成功返回的路径必须是完整产物, 且不带临时文件标记
func assertNoTempFiles(t *testing.T, dir string) {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read dir: %v", err)
	}
	for _, entry := range entries {
		if osutil.IsTempFile(entry.Name()) {
			t.Errorf("leftover temp file: %s", entry.Name())
		}
	}
}

func assertOnlyFiles(t *testing.T, dir string, names ...string) {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read dir: %v", err)
	}
	if len(entries) != len(names) {
		t.Fatalf("unexpected files in %s: %v", dir, entryNames(entries))
	}
	for i, entry := range entries {
		if entry.Name() != names[i] {
			t.Errorf("unexpected file %s, want %s", entry.Name(), names[i])
		}
	}
}

func entryNames(entries []os.DirEntry) []string {
	names := make([]string, len(entries))
	for i, entry := range entries {
		names[i] = entry.Name()
	}
	return names
}
