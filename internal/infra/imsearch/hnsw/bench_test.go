package hnsw

import (
	"math/rand"
	"testing"
)

func buildVecs(n, dim int, seed int64) [][]byte {
	rng := rand.New(rand.NewSource(seed))
	vecs := make([][]byte, n)
	for i := range vecs {
		vecs[i] = make([]byte, dim)
		rng.Read(vecs[i])
	}
	return vecs
}

func BenchmarkBuild(b *testing.B) {
	const n, dim = 4096, 32 // ORB centroid scale
	vecs := buildVecs(n, dim, 1)
	b.ReportAllocs()

	for b.Loop() {
		h := New(DefaultParams())
		for _, v := range vecs {
			h.Add(v)
		}
	}
}

func BenchmarkSearch(b *testing.B) {
	const n, dim = 4096, 32
	vecs := buildVecs(n, dim, 2)
	h := New(DefaultParams())
	for _, v := range vecs {
		h.Add(v)
	}
	queries := buildVecs(1000, dim, 3)
	b.ReportAllocs()
	b.ResetTimer()
	for i := range b.N {
		h.Search(queries[i%len(queries)], 1)
	}
}
