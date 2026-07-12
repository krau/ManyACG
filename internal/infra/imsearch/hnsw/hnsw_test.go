package hnsw

import (
	"math/rand"
	"path/filepath"
	"testing"

	"github.com/krau/ManyACG/internal/infra/imsearch/hamming"
)

func TestHNSWRecall(t *testing.T) {
	rng := rand.New(rand.NewSource(123))
	const n, dim = 2000, 64
	vecs := make([][]byte, n)
	for i := range vecs {
		vecs[i] = make([]byte, dim)
		rng.Read(vecs[i])
	}

	p := DefaultParams()
	p.EfSearch = 64 // higher ef for the recall check
	h := New(p)
	for _, v := range vecs {
		h.Add(v)
	}
	if h.Len() != n {
		t.Fatalf("Len=%d want %d", h.Len(), n)
	}

	// recall@10: fraction of queries whose true nearest neighbor appears in the
	// returned top-10. Pure-random 512-bit vectors are a hard case (distances
	// cluster near 256), so we verify the graph finds the true NN within a
	// small candidate set rather than exact top-1.
	hits := 0
	trials := 200
	for range trials {
		q := make([]byte, dim)
		rng.Read(q)
		got := h.Search(q, 10)

		best := uint32(1 << 30)
		var bestID uint32
		for i, v := range vecs {
			if d := hamming.Distance(q, v); d < best {
				best = d
				bestID = uint32(i)
			}
		}
		for _, r := range got {
			if r.ID == bestID {
				hits++
				break
			}
		}
	}
	recall := float64(hits) / float64(trials)
	if recall < 0.90 {
		t.Fatalf("recall@10 = %.2f, want >= 0.90", recall)
	}
	t.Logf("recall@10 = %.3f", recall)
}

func TestHNSWSaveLoad(t *testing.T) {
	rng := rand.New(rand.NewSource(7))
	const n, dim = 500, 64
	vecs := make([][]byte, n)
	for i := range vecs {
		vecs[i] = make([]byte, dim)
		rng.Read(vecs[i])
	}
	h := New(DefaultParams())
	for _, v := range vecs {
		h.Add(v)
	}

	path := filepath.Join(t.TempDir(), "q.bin")
	if err := h.Save(path); err != nil {
		t.Fatal(err)
	}
	h2, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if h2.Len() != n {
		t.Fatalf("loaded Len=%d want %d", h2.Len(), n)
	}
	// Same query must give identical results before/after round-trip.
	for range 50 {
		q := make([]byte, dim)
		rng.Read(q)
		r1 := h.Search(q, 5)
		r2 := h2.Search(q, 5)
		if len(r1) != len(r2) {
			t.Fatalf("result count differs: %d vs %d", len(r1), len(r2))
		}
		for i := range r1 {
			if r1[i].ID != r2[i].ID || r1[i].Distance != r2[i].Distance {
				t.Fatalf("result %d differs after reload", i)
			}
		}
	}
}

func TestHNSWVectorsPreserved(t *testing.T) {
	h := New(DefaultParams())
	v := []byte{1, 2, 3, 4, 5, 6, 7, 8}
	id := h.Add(v)
	v[0] = 99 // mutate caller's slice
	if h.Vectors()[id][0] != 1 {
		t.Fatal("Add did not copy the vector")
	}
}

func TestHNSWClusteredRecall(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	const nClusters, dim = 4096, 32
	const noiseBits = 3 // flip 3 random bits per query

	// Build cluster centers (like k-modes centroids).
	centers := make([][]byte, nClusters)
	for i := range centers {
		centers[i] = make([]byte, dim)
		rng.Read(centers[i])
	}

	h := New(DefaultParams())
	h.params.EfSearch = 64
	for _, c := range centers {
		h.Add(c)
	}

	hits := 0
	trials := 500
	for range trials {
		// Pick a random center, flip a few bits to create a query near it.
		target := rng.Intn(nClusters)
		q := make([]byte, dim)
		copy(q, centers[target])
		for range noiseBits {
			bit := rng.Intn(dim * 8)
			q[bit/8] ^= 1 << (bit % 8)
		}
		got := h.Search(q, 1)
		if len(got) > 0 && got[0].ID == uint32(target) {
			hits++
		}
	}
	recall := float64(hits) / float64(trials)
	t.Logf("clustered recall@1 = %.3f (nClusters=%d, noiseBits=%d)", recall, nClusters, noiseBits)
	if recall < 0.95 {
		t.Fatalf("clustered recall@1 = %.3f, want >= 0.95", recall)
	}
}
