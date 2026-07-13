package hnsw

import (
	"math/rand"
	"testing"

	"github.com/krau/ManyACG/internal/infra/imseek/hamming"
)

func TestHeuristicVsSimpleRecall(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	const nClusters, dim = 4096, 32
	const noiseBits = 8

	// Build centers that are closer together (not fully random) to create
	// ambiguous queries where graph quality matters. Start with 256 random
	// "seed" centers, then generate 16 variants of each by flipping 8 bits.
	// This gives 4096 centers in 256 tight clusters.
	centers := make([][]byte, nClusters)
	for i := range nClusters {
		seedIdx := i / 16
		if i%16 == 0 {
			centers[i] = make([]byte, dim)
			rng.Read(centers[i])
		} else {
			c := make([]byte, dim)
			copy(c, centers[seedIdx*16])
			for range 8 {
				bit := rng.Intn(dim * 8)
				c[bit/8] ^= 1 << (bit % 8)
			}
			centers[i] = c
		}
	}

	// Generate queries
	type query struct {
		q      []byte
		target uint32
	}
	queries := make([]query, 1000)
	for i := range queries {
		target := rng.Intn(nClusters)
		q := make([]byte, dim)
		copy(q, centers[target])
		for range noiseBits {
			bit := rng.Intn(dim * 8)
			q[bit/8] ^= 1 << (bit % 8)
		}
		queries[i] = query{q: q, target: uint32(target)}
	}

	// Build with heuristic (current implementation)
	h := New(DefaultParams())
	h.params.EfSearch = 8 // low ef to stress-test the graph quality
	for _, c := range centers {
		h.Add(c)
	}
	hitsHeuristic := 0
	for _, q := range queries {
		got := h.Search(q.q, 1)
		if len(got) > 0 && got[0].ID == q.target {
			hitsHeuristic++
		}
	}

	// Build with simple truncation for comparison
	h2 := New(DefaultParams())
	h2.params.EfSearch = 8
	h2.params.SimpleNeighborSelect = true
	for _, c := range centers {
		h2.Add(c)
	}
	hitsSimple := 0
	for _, q := range queries {
		got := h2.Search(q.q, 1)
		if len(got) > 0 && got[0].ID == q.target {
			hitsSimple++
		}
	}

	// Compute brute-force recall baseline (ground truth)
	hitsBrute := 0
	for _, q := range queries {
		best := uint32(1 << 30)
		var bestID uint32
		for i, c := range centers {
			if d := hamming.Distance(q.q, c); d < best {
				best = d
				bestID = uint32(i)
			}
		}
		if bestID == q.target {
			hitsBrute++
		}
	}

	t.Logf("heuristic recall@1 = %.3f (%d/%d)",
		float64(hitsHeuristic)/float64(len(queries)), hitsHeuristic, len(queries))
	t.Logf("simple    recall@1 = %.3f (%d/%d)",
		float64(hitsSimple)/float64(len(queries)), hitsSimple, len(queries))
	t.Logf("brute-force target hit rate = %.3f (%d/%d) [confirms queries are valid]",
		float64(hitsBrute)/float64(len(queries)), hitsBrute, len(queries))

	if float64(hitsHeuristic)/float64(len(queries)) < 0.90 {
		t.Fatalf("heuristic recall@1 = %.3f, want >= 0.90", float64(hitsHeuristic)/float64(len(queries)))
	}
}
