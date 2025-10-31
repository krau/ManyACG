package kvstor

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/krau/ManyACG/internal/shared/errs"
	"go.etcd.io/bbolt"
)

// helper to create a temporary bboltDB instance for tests
func newTestBboltDB(t *testing.T) *bboltDB {
	t.Helper()
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.kv.bbolt")
	db, err := bbolt.Open(dbPath, 0600, nil)
	if err != nil {
		t.Fatalf("open bbolt: %v", err)
	}
	t.Cleanup(func() {
		db.Close()
		os.Remove(dbPath)
	})
	return &bboltDB{
		db:             db,
		bucket:         "test_bucket",
		ttlBucket:      "test_bucket_ttl",
		ttlBatchLimit:  1024,
		ttlSweepPeriod: 50 * time.Millisecond,
		stop:           make(chan struct{}),
	}
}

func TestBboltSetGet_NoTTL(t *testing.T) {
	b := newTestBboltDB(t)
	ctx := context.Background()

	key := "k1"
	val := "v1"
	if err := b.Set(ctx, key, val, 0); err != nil {
		t.Fatalf("Set failed: %v", err)
	}
	got, err := b.Get(ctx, key)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	s, ok := got.(string)
	if !ok || s != val {
		t.Fatalf("Get value mismatch: got=%T %v want=%q", got, got, val)
	}
}

func TestBboltDelete(t *testing.T) {
	b := newTestBboltDB(t)
	ctx := context.Background()

	key := "kdel"
	if err := b.Set(ctx, key, "val", 0); err != nil {
		t.Fatalf("Set failed: %v", err)
	}
	if err := b.Delete(ctx, key); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
	if _, err := b.Get(ctx, key); err == nil || err != errs.ErrRecordNotFound {
		t.Fatalf("expected ErrRecordNotFound after delete, got: %v", err)
	}
}

func TestBboltTTL_ExpireOnGet(t *testing.T) {
	b := newTestBboltDB(t)
	ctx := context.Background()

	key := "k_ttl_get"
	if err := b.Set(ctx, key, "v", 30*time.Millisecond); err != nil {
		t.Fatalf("Set with TTL failed: %v", err)
	}
	time.Sleep(70 * time.Millisecond)
	if _, err := b.Get(ctx, key); err == nil || err != errs.ErrRecordNotFound {
		t.Fatalf("expected ErrRecordNotFound after TTL, got: %v", err)
	}
}

func TestBboltTTL_SweepExpired(t *testing.T) {
	b := newTestBboltDB(t)
	ctx := context.Background()

	// set three keys with short TTL
	keys := []string{"k1", "k2", "k3"}
	for _, k := range keys {
		if err := b.Set(ctx, k, "v", 20*time.Millisecond); err != nil {
			t.Fatalf("Set with TTL failed: %v", err)
		}
	}
	// wait for expiration
	time.Sleep(60 * time.Millisecond)
	// sweep explicitly
	if err := b.sweepExpired(); err != nil {
		t.Fatalf("sweepExpired failed: %v", err)
	}

	// verify keys are deleted at storage level
	err := b.db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte(b.bucket))
		if bucket == nil {
			return nil
		}
		for _, k := range keys {
			if v := bucket.Get([]byte(k)); v != nil {
				t.Fatalf("key %s should have been deleted by sweep", k)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("view tx failed: %v", err)
	}
}

func TestBboltTTL_Overwrite(t *testing.T) {
	b := newTestBboltDB(t)
	ctx := context.Background()

	key := "k_overwrite"
	// set with long ttl
	if err := b.Set(ctx, key, "v", time.Second); err != nil {
		t.Fatalf("Set failed: %v", err)
	}
	// overwrite with short ttl
	if err := b.Set(ctx, key, "v2", 20*time.Millisecond); err != nil {
		t.Fatalf("Set overwrite failed: %v", err)
	}
	// immediate get returns new value
	got, err := b.Get(ctx, key)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if s := got.(string); s != "v2" {
		t.Fatalf("value mismatch after overwrite, got=%q want=%q", s, "v2")
	}
	// after ttl passed, key should be gone
	time.Sleep(60 * time.Millisecond)
	if _, err := b.Get(ctx, key); err == nil || err != errs.ErrRecordNotFound {
		t.Fatalf("expected not found after overwrite ttl, got=%v", err)
	}
}
