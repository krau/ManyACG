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
	}
}

func TestCompressImgFailureWritesNothing(t *testing.T) {
	dir := t.TempDir()
	inputPath := filepath.Join(dir, "input.png")
	writeTestPNG(t, inputPath, 32, 32)
	if _, err := CompressImg(inputPath, filepath.Join(dir, "output.unknown"), "unknown-format", 0); err == nil {
		t.Fatal("expected an error for unsupported format")
	}
	assertOnlyFiles(t, dir, "input.png")
}

// ugoira 转 mp4 的输出是临时产物, 并发转换同一目标时不应写入同一路径
func TestUgoiraZipToMp4UsesUniqueOutput(t *testing.T) {
	if !FFmpegAvailable() {
		t.Skip("ffmpeg is not available")
	}
	dir := t.TempDir()
	zipPath := filepath.Join(dir, "frames.zip")
	frames := writeTestUgoiraZip(t, zipPath, 64)
	basePath := filepath.Join(dir, "out.mp4")

	const workers = 3
	paths := make([]string, workers)
	errs := make([]error, workers)
	var wg sync.WaitGroup
	for i := range workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			paths[i], errs[i] = UgoiraZipToMp4(zipPath, frames, basePath)
		}()
	}
	wg.Wait()

	seen := make(map[string]struct{}, workers)
	for i := range workers {
		if errs[i] != nil {
			t.Fatalf("UgoiraZipToMp4: %v", errs[i])
		}
		if paths[i] == basePath {
			t.Fatalf("output is written to the shared path %s", paths[i])
		}
		if _, ok := seen[paths[i]]; ok {
			t.Fatalf("duplicated output path %s", paths[i])
		}
		seen[paths[i]] = struct{}{}
		if filepath.Ext(paths[i]) != ".mp4" {
			t.Errorf("unexpected output extension: %s", paths[i])
		}
		f, err := os.Open(paths[i])
		if err != nil {
			t.Fatalf("open output: %v", err)
		}
		meta, err := GetMP4Meta(f)
		f.Close()
		if err != nil {
			t.Fatalf("output %s is not a valid mp4: %v", paths[i], err)
		}
		if meta.Width != 64 || meta.Height != 64 || meta.Duration == 0 {
			t.Errorf("unexpected mp4 metadata for %s: %+v", paths[i], meta)
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
