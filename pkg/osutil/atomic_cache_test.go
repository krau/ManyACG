package osutil

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestAtomicCachePublishesCompleteFileOnly(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "cache.bin")
	payload := bytes.Repeat([]byte("atomic-cache"), 512)

	writer, err := CreateAtomicCache(target)
	if err != nil {
		t.Fatalf("CreateAtomicCache: %v", err)
	}
	if _, err := writer.Write(payload[:100]); err != nil {
		t.Fatalf("Write: %v", err)
	}
	if _, err := os.Stat(target); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("target must not exist before Commit: %v", err)
	}
	if _, err := writer.Write(payload[100:]); err != nil {
		t.Fatalf("Write: %v", err)
	}
	file, err := writer.Commit()
	if err != nil {
		t.Fatalf("Commit: %v", err)
	}
	defer file.Close()
	if _, err := writer.Write(payload); !errors.Is(err, os.ErrClosed) {
		t.Errorf("Write after Commit should fail with os.ErrClosed, got %v", err)
	}
	got, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("read target: %v", err)
	}
	if !bytes.Equal(got, payload) {
		t.Errorf("content mismatch: got %d bytes, want %d", len(got), len(payload))
	}
	if file.Name() != target {
		t.Errorf("unexpected file name: %s", file.Name())
	}
	assertNoTempFiles(t, dir)
}

func TestAtomicCacheAbortLeavesNothing(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "cache.bin")
	writer, err := CreateAtomicCache(target)
	if err != nil {
		t.Fatalf("CreateAtomicCache: %v", err)
	}
	if _, err := writer.Write(bytes.Repeat([]byte("x"), 64)); err != nil {
		t.Fatalf("Write: %v", err)
	}
	writer.Abort()
	writer.Abort() // 重复调用不应 panic
	if _, err := writer.Write([]byte("x")); !errors.Is(err, os.ErrClosed) {
		t.Errorf("Write after Abort should fail with os.ErrClosed, got %v", err)
	}
	if _, err := os.Stat(target); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("target must not exist after Abort: %v", err)
	}
	assertNoTempFiles(t, dir)
}
