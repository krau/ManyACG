package httpclient

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/imroc/req/v3"
)

func TestDownloadToFileIsAtomic(t *testing.T) {
	payload := bytes.Repeat([]byte("no-partial-files-in-cache"), 4096)
	started := make(chan struct{})
	release := make(chan struct{})
	var startOnce, releaseOnce sync.Once
	var srv *httptest.Server
	releaseNow := func() {
		releaseOnce.Do(func() { close(release) })
	}
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", strconv.Itoa(len(payload)))
		half := len(payload) / 2
		if _, err := w.Write(payload[:half]); err != nil {
			return
		}
		w.(http.Flusher).Flush()
		startOnce.Do(func() { close(started) })
		select {
		case <-release:
		case <-r.Context().Done():
			return
		}
		_, _ = w.Write(payload[half:])
	}))
	defer srv.Close()
	// 保证测试提前结束时服务端不会被阻塞在 release 上, 否则 Close 会一直等待
	defer releaseNow()

	dir := t.TempDir()
	path := filepath.Join(dir, "data.bin")
	errCh := make(chan error, 1)
	go func() {
		errCh <- downloadToFile(context.Background(), req.C(), srv.URL+"/slow.bin", path)
	}()

	<-started
	// 等待客户端开始写入(临时文件出现, 或旧实现直接写入了目标文件)
	deadline := time.Now().Add(5 * time.Second)
	for {
		if time.Now().After(deadline) {
			t.Fatal("download did not start writing")
		}
		entries, err := os.ReadDir(dir)
		if err != nil {
			t.Fatalf("read dir: %v", err)
		}
		wrote := len(entries) > 0
		if !wrote {
			time.Sleep(time.Millisecond)
			continue
		}
		break
	}
	// 服务端此时还未发送剩下的内容, 目标文件不应可读
	if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("target file is readable before the download completed: %v", err)
	}
	releaseNow()
	if err := <-errCh; err != nil {
		t.Fatalf("downloadToFile: %v", err)
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read target file: %v", err)
	}
	if !bytes.Equal(got, payload) {
		t.Errorf("target file content mismatch: got %d bytes, want %d", len(got), len(payload))
	}
	assertNoTempFiles(t, dir)
}

func TestDownloadToFileFailureLeavesNoFile(t *testing.T) {
	payload := bytes.Repeat([]byte("truncated-body"), 128)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/broken.bin" {
			// 声明的内容长度大于实际写入的字节, 提前关闭连接
			w.Header().Set("Content-Length", strconv.Itoa(len(payload)*2))
			_, _ = w.Write(payload)
			return
		}
		http.Error(w, "not found", http.StatusNotFound)
	}))
	defer srv.Close()

	for _, tt := range []struct {
		name string
		url  string
	}{
		{name: "truncated response", url: srv.URL + "/broken.bin"},
		{name: "error status", url: srv.URL + "/missing.bin"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			path := filepath.Join(dir, "data.bin")
			if err := downloadToFile(context.Background(), req.C(), tt.url, path); err == nil {
				t.Fatal("expected an error")
			} else if tt.name == "error status" && !strings.Contains(err.Error(), fmt.Sprint(http.StatusNotFound)) {
				t.Errorf("error should mention the status code, got: %v", err)
			}
			if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
				t.Errorf("target file must not exist after a failed download: %v", err)
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
