package osutil

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func assertNoTempFiles(t *testing.T, dir string) {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read dir: %v", err)
	}
	for _, entry := range entries {
		if IsTempFile(entry.Name()) {
			t.Errorf("leftover temp file: %s", entry.Name())
		}
	}
}

func TestCreateTempSiblingKeepsExtension(t *testing.T) {
	dir := t.TempDir()
	f, err := CreateTempSibling(filepath.Join(dir, "output.mp4"))
	if err != nil {
		t.Fatalf("CreateTempSibling: %v", err)
	}
	defer f.Close()
	if !IsTempFile(f.Name()) {
		t.Errorf("temp file name does not carry the prefix: %s", f.Name())
	}
	if filepath.Ext(f.Name()) != ".mp4" {
		t.Errorf("extension is not preserved: %s", f.Name())
	}
	if filepath.Dir(f.Name()) != dir {
		t.Errorf("temp file is not a sibling: %s", f.Name())
	}
	if p := TempSiblingPath(filepath.Join(dir, "output.png")); !IsTempFile(p) || filepath.Ext(p) != ".png" {
		t.Errorf("unexpected TempSiblingPath: %s", p)
	}
}

func TestUniquePathDoesNotRepeat(t *testing.T) {
	dir := t.TempDir()
	first := UniquePath(filepath.Join(dir, "out.jpg"))
	second := UniquePath(filepath.Join(dir, "out.jpg"))
	if first == second {
		t.Errorf("UniquePath returned the same path twice: %s", first)
	}
	for _, p := range []string{first, second} {
		if IsTempFile(p) {
			t.Errorf("unique path must not look like a temp file: %s", p)
		}
		if filepath.Ext(p) != ".jpg" || filepath.Dir(p) != dir {
			t.Errorf("unexpected path: %s", p)
		}
	}
}

func TestRemoveStaleTempFiles(t *testing.T) {
	dir := t.TempDir()
	subDir := filepath.Join(dir, "req")
	if err := os.MkdirAll(subDir, 0o755); err != nil {
		t.Fatal(err)
	}
	newTemp := func(d string, target string) string {
		f, err := CreateTempSibling(filepath.Join(d, target))
		if err != nil {
			t.Fatal(err)
		}
		f.Close()
		return f.Name()
	}
	staleTop := newTemp(dir, "a.jpg")
	staleSub := newTemp(subDir, "b.jpg")
	fresh := newTemp(dir, "c.jpg")
	keep := filepath.Join(dir, "d5f1c7.jpg")
	if err := os.WriteFile(keep, []byte("cached"), 0o644); err != nil {
		t.Fatal(err)
	}
	old := time.Now().Add(-2 * time.Hour)
	for _, p := range []string{staleTop, staleSub} {
		if err := os.Chtimes(p, old, old); err != nil {
			t.Fatal(err)
		}
	}

	removed, err := RemoveStaleTempFiles(dir, time.Hour)
	if err != nil {
		t.Fatalf("RemoveStaleTempFiles: %v", err)
	}
	if removed != 2 {
		t.Errorf("removed %d files, want 2", removed)
	}
	for _, p := range []string{staleTop, staleSub} {
		if _, err := os.Stat(p); !errors.Is(err, os.ErrNotExist) {
			t.Errorf("stale temp file still exists: %s", p)
		}
	}
	for _, p := range []string{fresh, keep} {
		if _, err := os.Stat(p); err != nil {
			t.Errorf("file should be kept: %s: %v", p, err)
		}
	}

	removed, err = RemoveStaleTempFiles(filepath.Join(dir, "missing"), time.Hour)
	if err != nil || removed != 0 {
		t.Errorf("missing dir: removed=%d err=%v", removed, err)
	}
}
