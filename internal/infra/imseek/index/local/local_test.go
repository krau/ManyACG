package local

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/krau/ManyACG/internal/infra/imseek/index"
)

func TestNew(t *testing.T) {
	b := New("/tmp/test", 32)
	if b == nil {
		t.Fatal("New returned nil")
	}
	if b.codeSize != 32 {
		t.Errorf("codeSize = %d want 32", b.codeSize)
	}
}

func TestTrained_EmptyDir(t *testing.T) {
	ctx := context.Background()
	b := New(t.TempDir(), 32)
	if b.Trained(ctx) {
		t.Error("Trained() = true for empty dir, want false")
	}
}

func TestTrained_QuantizerExists(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	// Create a fake quantizer file
	if err := os.WriteFile(filepath.Join(dir, "quantizer.bin"), []byte("fake"), 0o644); err != nil {
		t.Fatal(err)
	}
	b := New(dir, 32)
	if !b.Trained(ctx) {
		t.Error("Trained() = false with quantizer.bin present, want true")
	}
}

func TestTrain_EmptySamples(t *testing.T) {
	ctx := context.Background()
	b := New(t.TempDir(), 32)
	_, err := b.Train(ctx, nil, index.TrainOptions{NList: 64})
	if err == nil {
		t.Error("Train with nil samples should error")
	}
}

func TestTrain_ZeroNList(t *testing.T) {
	ctx := context.Background()
	b := New(t.TempDir(), 32)
	samples := make([][]byte, 10)
	for i := range samples {
		samples[i] = make([]byte, 32)
	}
	_, err := b.Train(ctx, samples, index.TrainOptions{NList: 0})
	if err == nil {
		t.Error("Train with NList=0 should error")
	}
}

func TestTrainAndSearchRoundTrip(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	b := New(dir, 8) // 8-byte codes for speed

	// Create synthetic training vectors: 4 clusters of 10 vectors each
	const nClusters = 4
	const perCluster = 10
	samples := make([][]byte, 0, nClusters*perCluster)
	for c := range nClusters {
		for range perCluster {
			v := make([]byte, 8)
			v[0] = byte(c * 50) // distinct first byte per cluster
			samples = append(samples, v)
		}
	}

	res, err := b.Train(ctx, samples, index.TrainOptions{
		NList:    nClusters,
		MaxIter:  5,
		Init:     "kmeans-plus-plus",
		TwoLevel: false,
	})
	if err != nil {
		t.Fatalf("Train: %v", err)
	}
	if res.Centroids != nClusters {
		t.Errorf("Centroids = %d want %d", res.Centroids, nClusters)
	}
	if !b.Trained(ctx) {
		t.Error("Trained() = false after Train")
	}

	// Build: insert all vectors
	builder, err := b.NewBuilder(ctx)
	if err != nil {
		t.Fatalf("NewBuilder: %v", err)
	}
	for i, s := range samples {
		if err := builder.Add(ctx, [][]byte{s}, []uint64{uint64(i + 1)}); err != nil {
			t.Fatalf("Add(%d): %v", i, err)
		}
	}
	if err := builder.Flush(ctx); err != nil {
		t.Fatalf("Flush: %v", err)
	}
	if err := builder.Commit(ctx); err != nil {
		t.Fatalf("Commit: %v", err)
	}
	builder.Close()

	// Search: query a vector close to cluster 0
	searcher, err := b.NewSearcher(ctx, 1)
	if err != nil {
		t.Fatalf("NewSearcher: %v", err)
	}
	defer searcher.Close()

	query := make([]byte, 8)
	query[0] = 0 // close to cluster 0
	results := searcher.Search(ctx, [][]byte{query}, 3, 1)
	if len(results) == 0 {
		t.Fatal("Search returned no results")
	}
	// Results should be from cluster 0 (ids 1-10)
	for _, r := range results {
		if r.ID < 1 || r.ID > 10 {
			t.Errorf("result ID = %d, expected 1-10 (cluster 0)", r.ID)
		}
	}
}

func TestClose(t *testing.T) {
	b := New(t.TempDir(), 32)
	if err := b.Close(); err != nil {
		t.Errorf("Close() = %v", err)
	}
}
