package utils

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/krau/ManyACG/pkg/osutil"
	"github.com/krau/ManyACG/pkg/strutil"
)

var corruptPictureFormatID atomic.Uint64

type pictureWaitContext struct {
	context.Context
	waiting chan struct{}
	once    sync.Once
}

func (c *pictureWaitContext) Done() <-chan struct{} {
	c.once.Do(func() { close(c.waiting) })
	return c.Context.Done()
}

func TestDownloadAndDecodePictureConcurrentRecovery(t *testing.T) {
	if os.Getenv("MANYACG_PICTURE_RECOVERY_TEST") != "1" {
		executable, err := os.Executable()
		if err != nil {
			t.Fatal(err)
		}
		cmd := exec.Command(executable, "-test.run=^TestDownloadAndDecodePictureConcurrentRecovery$", "-test.timeout=20s")
		cmd.Env = append(os.Environ(), "MANYACG_PICTURE_RECOVERY_TEST=1")
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("isolated picture recovery: %v\n%s", err, out)
		}
		return
	}
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "config.toml"), []byte(fmt.Sprintf("[storage]\ncache_dir = %q\n", dir)), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Chdir(dir)

	firstDecode, secondDecode := make(chan struct{}), make(chan struct{})
	releaseFirst, releaseSecond := make(chan struct{}), make(chan struct{})
	var firstOnce, secondOnce sync.Once
	defer firstOnce.Do(func() { close(releaseFirst) })
	defer secondOnce.Do(func() { close(releaseSecond) })
	magic := fmt.Sprintf("ManyACGCorrupt%d#", corruptPictureFormatID.Add(1))
	var decodes atomic.Int32
	image.RegisterFormat(magic, magic, func(io.Reader) (image.Image, error) {
		switch decodes.Add(1) {
		case 1:
			close(firstDecode)
			<-releaseFirst
		case 2:
			close(secondDecode)
			<-releaseSecond
		}
		return nil, errors.New("corrupt picture")
	}, func(io.Reader) (image.Config, error) {
		return image.Config{}, errors.New("corrupt picture")
	})

	var payload bytes.Buffer
	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	img.Set(0, 0, color.RGBA{R: 255, A: 255})
	if err := png.Encode(&payload, img); err != nil {
		t.Fatal(err)
	}
	redownloadStarted, releaseRedownload := make(chan struct{}), make(chan struct{})
	var downloadOnce sync.Once
	var requests atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/picture.png" && requests.Add(1) == 2 {
			close(redownloadStarted)
			select {
			case <-releaseRedownload:
			case <-r.Context().Done():
				return
			}
		}
		w.Header().Set("Content-Type", "image/png")
		_, _ = w.Write(payload.Bytes())
	}))
	defer srv.Close()
	defer downloadOnce.Do(func() { close(releaseRedownload) })
	url := srv.URL + "/picture.png"
	path := filepath.Join(dir, "req", strutil.MD5Hash(url)+".png")
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(magic), 0o600); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	type result struct {
		file *osutil.File
		err  error
	}
	call := func(callCtx context.Context, ch chan result) {
		f, _, err := downloadAndDecodePicture(callCtx, url)
		ch <- result{f, err}
	}
	firstResult, secondResult := make(chan result, 1), make(chan result, 1)
	go call(ctx, firstResult)
	select {
	case <-firstDecode:
	case <-ctx.Done():
		t.Fatal("first decoder did not start")
	}

	other, _, err := downloadAndDecodePicture(ctx, srv.URL+"/other.png")
	if err != nil {
		t.Fatalf("unrelated URL blocked by recovery: %v", err)
	}
	other.Close()

	waitCtx := &pictureWaitContext{Context: ctx, waiting: make(chan struct{})}
	go call(waitCtx, secondResult)
	select {
	case <-waitCtx.waiting:
	case <-secondDecode:
	case <-ctx.Done():
		t.Fatal("second caller did not reach the cache")
	}
	firstOnce.Do(func() { close(releaseFirst) })
	var good result
	select {
	case good = <-firstResult:
	case <-ctx.Done():
		t.Fatal("first recovery did not complete")
	}
	if good.err != nil {
		t.Fatal(good.err)
	}
	defer good.file.Close()
	secondOnce.Do(func() { close(releaseSecond) })
	var second result
	select {
	case <-redownloadStarted:
		t.Fatal("stale decoder invalidated the recovered cache")
	case second = <-secondResult:
	case <-ctx.Done():
		t.Fatal("second recovery did not complete")
	}
	if second.err != nil {
		t.Fatal(second.err)
	}
	defer second.file.Close()
	got, err := os.ReadFile(good.file.Name())
	if err != nil {
		t.Fatalf("returned picture cannot be reopened: %v", err)
	}
	if !bytes.Equal(got, payload.Bytes()) {
		t.Fatal("recovered picture content mismatch")
	}
	if requests.Load() != 1 {
		t.Fatalf("same URL downloaded %d times, want 1", requests.Load())
	}
	pictureCacheMu.Lock()
	defer pictureCacheMu.Unlock()
	if len(pictureCacheLocks) != 0 {
		t.Fatalf("unused picture locks retained: %d", len(pictureCacheLocks))
	}
}

func TestPictureCacheWaitCancellation(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	lock, err := acquirePictureCacheLock(ctx, "cancelled-picture")
	if err != nil {
		t.Fatal(err)
	}
	defer releasePictureCacheLock("cancelled-picture", lock)
	waitingCtx, cancelWaiting := context.WithCancel(ctx)
	defer cancelWaiting()
	waitCtx := &pictureWaitContext{Context: waitingCtx, waiting: make(chan struct{})}
	done := make(chan error, 1)
	go func() {
		waitingLock, err := acquirePictureCacheLock(waitCtx, "cancelled-picture")
		if err == nil {
			releasePictureCacheLock("cancelled-picture", waitingLock)
		}
		done <- err
	}()
	<-waitCtx.waiting
	cancelWaiting()
	if err := <-done; !errors.Is(err, context.Canceled) {
		t.Fatalf("waiting caller returned %v, want context.Canceled", err)
	}
}
