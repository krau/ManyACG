package service

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/krau/ManyACG/internal/infra/config/runtimecfg"
	"github.com/krau/ManyACG/internal/infra/storage"
	"github.com/krau/ManyACG/internal/shared"
	"github.com/krau/ManyACG/pkg/osutil"
)

// gatedReader 先返回前半部分数据, 阻塞直到 release 关闭后再返回剩余数据
type gatedReader struct {
	data    []byte
	offset  int
	half    int
	reached chan struct{}
	release chan struct{}
	once    sync.Once
	err     error
}

func (r *gatedReader) Read(p []byte) (int, error) {
	if r.offset < r.half {
		n := copy(p, r.data[r.offset:r.half])
		r.offset += n
		if r.offset >= r.half {
			r.once.Do(func() { close(r.reached) })
		}
		return n, nil
	}
	<-r.release
	if r.err != nil {
		return 0, r.err
	}
	if r.offset >= len(r.data) {
		return 0, io.EOF
	}
	n := copy(p, r.data[r.offset:])
	r.offset += n
	return n, nil
}

func (r *gatedReader) Close() error { return nil }

type stubStorage struct {
	reader io.ReadCloser
}

func (s *stubStorage) Init(context.Context) error { return nil }

func (s *stubStorage) Save(context.Context, io.Reader, string) (*shared.StorageDetail, error) {
	return nil, errors.New("not implemented")
}

func (s *stubStorage) GetFile(context.Context, shared.StorageDetail) (io.ReadCloser, error) {
	return s.reader, nil
}

func (s *stubStorage) Delete(context.Context, shared.StorageDetail) error { return nil }

func newStorageService(dir string, reader io.ReadCloser) (*Service, shared.StorageDetail) {
	serv := NewService(nil, nil, nil,
		map[shared.StorageType]storage.Storage{shared.StorageType("stub"): &stubStorage{reader: reader}},
		nil, runtimecfg.StorageConfig{CacheDir: dir})
	detail := shared.StorageDetail{Type: "stub", Path: "test/file.jpg", Mime: "image/jpeg"}
	return serv, detail
}

func TestStorageGetFileWritesCacheAtomically(t *testing.T) {
	payload := bytes.Repeat([]byte("storage-cache"), 8192)
	reader := &gatedReader{data: payload, half: len(payload) / 2, reached: make(chan struct{}), release: make(chan struct{})}
	dir := t.TempDir()
	serv, detail := newStorageService(dir, reader)
	cachePath := serv.storageCachePath(detail)

	type result struct {
		file *osutil.File
		err  error
	}
	resCh := make(chan result, 1)
	go func() {
		file, err := serv.StorageGetFile(context.Background(), detail)
		resCh <- result{file: file, err: err}
	}()

	<-reader.reached
	if _, err := os.Stat(cachePath); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("cache file is readable before the copy finished: %v", err)
	}
	close(reader.release)
	res := <-resCh
	if res.err != nil {
		t.Fatalf("StorageGetFile: %v", res.err)
	}
	defer res.file.Close()
	got, err := io.ReadAll(res.file)
	if err != nil {
		t.Fatalf("read cached file: %v", err)
	}
	if !bytes.Equal(got, payload) {
		t.Errorf("cached file content mismatch: got %d bytes, want %d", len(got), len(payload))
	}
	onDisk, err := os.ReadFile(cachePath)
	if err != nil {
		t.Fatalf("read cache file: %v", err)
	}
	if !bytes.Equal(onDisk, payload) {
		t.Errorf("cache file content mismatch: got %d bytes, want %d", len(onDisk), len(payload))
	}
	assertNoTempFiles(t, dir)
}

func TestStorageStreamFileWritesCacheAtomically(t *testing.T) {
	payload := bytes.Repeat([]byte("storage-stream"), 8192)
	reader := &gatedReader{data: payload, half: len(payload) / 2, reached: make(chan struct{}), release: make(chan struct{})}
	dir := t.TempDir()
	serv, detail := newStorageService(dir, reader)
	cachePath := serv.storageCachePath(detail)

	var out bytes.Buffer
	done := make(chan error, 1)
	go func() {
		done <- serv.StorageStreamFile(context.Background(), detail, &out)
	}()

	<-reader.reached
	if _, err := os.Stat(cachePath); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("cache file is readable before the copy finished: %v", err)
	}
	close(reader.release)
	if err := <-done; err != nil {
		t.Fatalf("StorageStreamFile: %v", err)
	}
	if !bytes.Equal(out.Bytes(), payload) {
		t.Errorf("streamed content mismatch: got %d bytes, want %d", out.Len(), len(payload))
	}
	onDisk, err := os.ReadFile(cachePath)
	if err != nil {
		t.Fatalf("read cache file: %v", err)
	}
	if !bytes.Equal(onDisk, payload) {
		t.Errorf("cache file content mismatch: got %d bytes, want %d", len(onDisk), len(payload))
	}
	assertNoTempFiles(t, dir)
}

func TestStorageFileCacheDiscardedOnFailure(t *testing.T) {
	payload := bytes.Repeat([]byte("storage-failure"), 8192)
	for _, stream := range []bool{false, true} {
		name := "GetFile"
		if stream {
			name = "StreamFile"
		}
		t.Run(name, func(t *testing.T) {
			reader := &gatedReader{
				data: payload, half: len(payload) / 2,
				reached: make(chan struct{}), release: make(chan struct{}),
				err: errors.New("storage read failed"),
			}
			dir := t.TempDir()
			serv, detail := newStorageService(dir, reader)
			cachePath := serv.storageCachePath(detail)

			done := make(chan error, 1)
			go func() {
				if stream {
					done <- serv.StorageStreamFile(context.Background(), detail, io.Discard)
					return
				}
				_, err := serv.StorageGetFile(context.Background(), detail)
				done <- err
			}()
			<-reader.reached
			close(reader.release)
			if err := <-done; err == nil {
				t.Fatal("expected an error")
			}
			if _, err := os.Stat(cachePath); !errors.Is(err, os.ErrNotExist) {
				t.Errorf("cache file must not exist after a failed copy: %v", err)
			}
			assertNoTempFiles(t, dir)
		})
	}
}

func assertNoTempFiles(t *testing.T, dir string) {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read dir: %v", err)
	}
	for _, entry := range entries {
		if strings.Contains(entry.Name(), ".tmp-") {
			t.Errorf("leftover temp file: %s", entry.Name())
		}
	}
}
