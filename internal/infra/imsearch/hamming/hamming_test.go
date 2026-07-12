package hamming

import (
	"math/rand"
	"slices"
	"sort"
	"testing"
)

func naiveDistance(a, b []byte) uint32 {
	var sum uint32
	for i := range a {
		x := a[i] ^ b[i]
		for x != 0 {
			sum += uint32(x & 1)
			x >>= 1
		}
	}
	return sum
}

func TestDistanceKnownValues(t *testing.T) {
	cases := []struct {
		a, b []byte
		want uint32
	}{
		{make([]byte, 64), make([]byte, 64), 0},
		{[]byte{0xFF}, []byte{0x00}, 8},
		{[]byte{0x0F, 0xF0}, []byte{0x00, 0x00}, 8},
		{[]byte{0xAA, 0x55}, []byte{0x55, 0xAA}, 16},
	}
	for i, c := range cases {
		if got := Distance(c.a, c.b); got != c.want {
			t.Errorf("case %d: Distance=%d want %d", i, got, c.want)
		}
	}
}

func TestDistanceMatchesNaive(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	for _, n := range []int{1, 7, 8, 15, 32, 64, 100} {
		for range 200 {
			a := make([]byte, n)
			b := make([]byte, n)
			rng.Read(a)
			rng.Read(b)
			if got, want := Distance(a, b), naiveDistance(a, b); got != want {
				t.Fatalf("n=%d: Distance=%d naive=%d", n, got, want)
			}
		}
	}
}

func TestDistanceSymmetric(t *testing.T) {
	rng := rand.New(rand.NewSource(1))
	a := make([]byte, 64)
	b := make([]byte, 64)
	rng.Read(a)
	rng.Read(b)
	if Distance(a, b) != Distance(b, a) {
		t.Fatal("distance not symmetric")
	}
}

func TestKNN(t *testing.T) {
	rng := rand.New(rand.NewSource(7))
	const n, dim = 500, 64
	vb := make([][]byte, n)
	for i := range vb {
		vb[i] = make([]byte, dim)
		rng.Read(vb[i])
	}
	q := make([]byte, dim)
	rng.Read(q)

	k := 10
	got := KNN(q, vb, k)
	if len(got) != k {
		t.Fatalf("got %d neighbors want %d", len(got), k)
	}

	// Brute-force reference: compute all distances, sort, take k smallest.
	type nd struct {
		idx int
		d   uint32
	}
	all := make([]nd, n)
	for i, v := range vb {
		all[i] = nd{i, Distance(q, v)}
	}
	sort.Slice(all, func(i, j int) bool { return all[i].d < all[j].d })

	// The set of distances returned must equal the k smallest distances.
	wantDist := make([]uint32, k)
	for i := range k {
		wantDist[i] = all[i].d
	}
	gotDist := make([]uint32, k)
	for i, nb := range got {
		gotDist[i] = nb.Distance
		// verify the reported distance matches the descriptor at that index
		if Distance(q, vb[nb.Index]) != nb.Distance {
			t.Fatalf("neighbor distance mismatch at %d", i)
		}
	}
	slices.Sort(gotDist)
	for i := range wantDist {
		if gotDist[i] != wantDist[i] {
			t.Fatalf("knn distance set mismatch: got %v want %v", gotDist, wantDist)
		}
	}
}

func TestKNNFewerThanK(t *testing.T) {
	vb := [][]byte{{0x00}, {0x01}}
	got := KNN([]byte{0x00}, vb, 5)
	if len(got) != 2 {
		t.Fatalf("expected 2 results, got %d", len(got))
	}
}

func TestKNNEmpty(t *testing.T) {
	if KNN([]byte{0}, nil, 3) != nil {
		t.Fatal("expected nil for empty candidate set")
	}
}

func BenchmarkDistance64(b *testing.B) {
	rng := rand.New(rand.NewSource(1))
	x := make([]byte, 64)
	y := make([]byte, 64)
	rng.Read(x)
	rng.Read(y)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Distance(x, y)
	}
}
