package kvstor

import (
	"errors"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"go.etcd.io/bbolt"

	"github.com/krau/ManyACG/internal/shared/errs"
)

func setupTestDB(tb testing.TB) {
	tb.Helper()

	if stop := reaperStop; stop != nil {
		close(stop)
		reaperStop = nil
	}
	if defaultDb != nil && defaultDb.db != nil {
		_ = defaultDb.db.Close()
	}
	defaultDb = nil
	initOnce = sync.Once{}
	reaperOnce = sync.Once{}

	tempDir := tb.TempDir()
	dbPath := filepath.Join(tempDir, "kvstore.db")
	db, err := bbolt.Open(dbPath, 0600, &bbolt.Options{Timeout: time.Second})
	if err != nil {
		tb.Fatalf("failed to open test kvdb: %v", err)
	}

	defaultDb = &bboltDB{
		db:             db,
		bucket:         "test_bucket",
		ttlBucket:      "test_ttl_bucket",
		ttlBatchLimit:  10,
		ttlSweepPeriod: 10 * time.Millisecond,
	}

	tb.Cleanup(func() {
		if stop := reaperStop; stop != nil {
			close(stop)
			reaperStop = nil
		}
		if defaultDb != nil && defaultDb.db != nil {
			_ = defaultDb.db.Close()
		}
		defaultDb = nil
		initOnce = sync.Once{}
		reaperOnce = sync.Once{}
	})
}

func TestSetAndGet(t *testing.T) {
	setupTestDB(t)

	if err := Set("alpha", "omega"); err != nil {
		t.Fatalf("unexpected error from Set: %v", err)
	}

	got, err := Get[string]("alpha")
	if err != nil {
		t.Fatalf("unexpected error from Get: %v", err)
	}
	if got != "omega" {
		t.Fatalf("unexpected value from Get, want %q got %q", "omega", got)
	}
}

func TestSetWithTTLExpiration(t *testing.T) {
	setupTestDB(t)

	ttl := 30 * time.Millisecond
	if err := SetWithTTL("ephemeral", "value", ttl); err != nil {
		t.Fatalf("SetWithTTL returned error: %v", err)
	}

	time.Sleep(ttl + 50*time.Millisecond)

	if _, err := Get[string]("ephemeral"); !errors.Is(err, errs.ErrRecordNotFound) {
		t.Fatalf("expected ErrRecordNotFound after expiration, got %v", err)
	}

	var remainingValue bool
	var ttlEntries int
	if err := defaultDb.db.View(func(tx *bbolt.Tx) error {
		if bucket := tx.Bucket([]byte(defaultDb.bucket)); bucket != nil {
			if v := bucket.Get([]byte("ephemeral")); v != nil {
				remainingValue = true
			}
		}
		if ttlBucket := tx.Bucket([]byte(defaultDb.ttlBucket)); ttlBucket != nil {
			cursor := ttlBucket.Cursor()
			for k, _ := cursor.First(); k != nil; k, _ = cursor.Next() {
				ttlEntries++
			}
		}
		return nil
	}); err != nil {
		t.Fatalf("failed to inspect buckets: %v", err)
	}

	if remainingValue {
		t.Fatalf("expected expired key to be removed from primary bucket")
	}
	if ttlEntries != 0 {
		t.Fatalf("expected ttl bucket to be empty, found %d entries", ttlEntries)
	}
}

func TestSetWithTTLOverridesExistingTTL(t *testing.T) {
	setupTestDB(t)

	if err := SetWithTTL("override", "first", time.Minute); err != nil {
		t.Fatalf("first SetWithTTL returned error: %v", err)
	}
	if err := SetWithTTL("override", "second", 2*time.Minute); err != nil {
		t.Fatalf("second SetWithTTL returned error: %v", err)
	}

	got, err := Get[string]("override")
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}
	if got != "second" {
		t.Fatalf("unexpected value after override, want %q got %q", "second", got)
	}

	var ttlBucketMissing bool
	ttlEntryCount := countTTLEntries(t, &ttlBucketMissing)
	if ttlBucketMissing {
		t.Fatalf("ttl bucket was not created")
	}
	if ttlEntryCount != 1 {
		t.Fatalf("expected exactly one ttl entry after override, got %d", ttlEntryCount)
	}
}

func TestSweepExpiredHonorsBatchLimit(t *testing.T) {
	setupTestDB(t)

	defaultDb.ttlBatchLimit = 1
	ttl := 25 * time.Millisecond

	if err := SetWithTTL("k1", "v1", ttl); err != nil {
		t.Fatalf("SetWithTTL for k1 failed: %v", err)
	}
	if err := SetWithTTL("k2", "v2", ttl); err != nil {
		t.Fatalf("SetWithTTL for k2 failed: %v", err)
	}

	time.Sleep(ttl + 40*time.Millisecond)

	if err := sweepExpired(defaultDb); err != nil {
		t.Fatalf("first sweepExpired returned error: %v", err)
	}

	if keys := bucketKeys(t); len(keys) != 1 {
		t.Fatalf("expected 1 remaining key after limited sweep, got %d", len(keys))
	}
	if ttlEntries := countTTLEntries(t, nil); ttlEntries != 1 {
		t.Fatalf("expected 1 ttl entry remaining after limited sweep, got %d", ttlEntries)
	}

	if err := sweepExpired(defaultDb); err != nil {
		t.Fatalf("second sweepExpired returned error: %v", err)
	}

	if keys := bucketKeys(t); len(keys) != 0 {
		t.Fatalf("expected all keys to be removed after second sweep, got %d", len(keys))
	}
	if ttlEntries := countTTLEntries(t, nil); ttlEntries != 0 {
		t.Fatalf("expected ttl bucket to be empty after second sweep, got %d", ttlEntries)
	}
}

func bucketKeys(tb testing.TB) []string {
	tb.Helper()

	var keys []string
	if err := defaultDb.db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte(defaultDb.bucket))
		if bucket == nil {
			return nil
		}
		cursor := bucket.Cursor()
		for k, _ := cursor.First(); k != nil; k, _ = cursor.Next() {
			keys = append(keys, string(k))
		}
		return nil
	}); err != nil {
		tb.Fatalf("failed to read bucket keys: %v", err)
	}
	return keys
}

func countTTLEntries(tb testing.TB, bucketMissing *bool) int {
	tb.Helper()

	var count int
	if err := defaultDb.db.View(func(tx *bbolt.Tx) error {
		ttlBucket := tx.Bucket([]byte(defaultDb.ttlBucket))
		if ttlBucket == nil {
			if bucketMissing != nil {
				*bucketMissing = true
			}
			return nil
		}
		cursor := ttlBucket.Cursor()
		for k, _ := cursor.First(); k != nil; k, _ = cursor.Next() {
			count++
		}
		return nil
	}); err != nil {
		tb.Fatalf("failed to count ttl entries: %v", err)
	}
	return count
}

func BenchmarkSet(b *testing.B) {
	setupTestDB(b)

	key := "bench_set"
	value := "payload"

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := Set(key, value); err != nil {
			b.Fatalf("Set failed: %v", err)
		}
	}
}

func BenchmarkSetWithTTL(b *testing.B) {
	setupTestDB(b)

	key := "bench_ttl"
	value := "payload"
	ttl := time.Second

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := SetWithTTL(key, value, ttl); err != nil {
			b.Fatalf("SetWithTTL failed: %v", err)
		}
	}
}

func BenchmarkGet(b *testing.B) {
	setupTestDB(b)

	key := "bench_get"
	value := "payload"
	if err := Set(key, value); err != nil {
		b.Fatalf("preparing key failed: %v", err)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := Get[string](key); err != nil {
			b.Fatalf("Get failed: %v", err)
		}
	}
}
