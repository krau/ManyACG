package ivf

import (
	"math/rand"
	"testing"

	"github.com/krau/ManyACG/internal/infra/imsearch/hamming"
	"github.com/krau/ManyACG/internal/infra/imsearch/ivf/invlists"
	"github.com/krau/ManyACG/internal/infra/imsearch/ivf/quantizer"
)

func TestXORPreservesDistance(t *testing.T) {
	rng := rand.New(rand.NewSource(1))
	a := make([]byte, 8)
	b := make([]byte, 8)
	c := make([]byte, 8)
	rng.Read(a)
	rng.Read(b)
	rng.Read(c)
	// H(a,b) == H(a^c, b^c)
	if hamming.Distance(a, b) != hamming.Distance(xor(a, c), xor(b, c)) {
		t.Fatal("XOR did not preserve Hamming distance")
	}
}

func TestIVFAddSearch(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	const cs = 8

	// Build a small set of centroids.
	nlist := 32
	centroids := make([][]byte, nlist)
	for i := range centroids {
		centroids[i] = make([]byte, cs)
		rng.Read(centroids[i])
	}
	q := quantizer.New(centroids)

	// Database vectors: clusters around random centroids with small noise.
	const nDB = 2000
	db := make([][]byte, nDB)
	ids := make([]uint64, nDB)
	for i := range nDB {
		base := centroids[rng.Intn(nlist)]
		v := make([]byte, cs)
		copy(v, base)
		// flip a couple bits
		for range 2 {
			bit := rng.Intn(cs * 8)
			v[bit/8] ^= 1 << (bit % 8)
		}
		db[i] = v
		ids[i] = uint64(i + 1)
	}

	il := invlists.NewArray(nlist)
	idx := New(q, il, 4)
	idx.Add(db, ids)

	// All vectors should have been assigned somewhere.
	total := 0
	for i := range nlist {
		total += il.ListLen(i)
	}
	if total != nDB {
		t.Fatalf("assigned %d vectors want %d", total, nDB)
	}

	// Query with an exact copy of a known db vector; it should be found with
	// distance 0 (probe enough lists).
	target := 500
	res := idx.Search([][]byte{db[target]}, 5, 8)
	found := false
	for _, n := range res {
		if n.ID == ids[target] && n.Distance == 0 {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("exact query did not find target with distance 0 (got %d neighbors)", len(res))
	}
}

func TestIVFSearchViaOnDisk(t *testing.T) {
	rng := rand.New(rand.NewSource(7))
	const cs, nlist, nDB = 8, 16, 1000

	centroids := make([][]byte, nlist)
	for i := range centroids {
		centroids[i] = make([]byte, cs)
		rng.Read(centroids[i])
	}
	q := quantizer.New(centroids)

	db := make([][]byte, nDB)
	ids := make([]uint64, nDB)
	for i := range nDB {
		db[i] = make([]byte, cs)
		copy(db[i], centroids[rng.Intn(nlist)])
		bit := rng.Intn(cs * 8)
		db[i][bit/8] ^= 1 << (bit % 8)
		ids[i] = uint64(i + 1)
	}

	il := invlists.NewArray(nlist)
	New(q, il, 4).Add(db, ids)

	path := t.TempDir() + "/invlists.bin"
	if err := invlists.Save(il, path, cs, 3); err != nil {
		t.Fatal(err)
	}
	od, err := invlists.LoadOnDisk(path)
	if err != nil {
		t.Fatal(err)
	}
	defer od.Close()

	idx := New(q, od, 4)
	target := 123
	res := idx.Search([][]byte{db[target]}, 5, nlist)
	found := false
	for _, n := range res {
		if n.ID == ids[target] && n.Distance == 0 {
			found = true
		}
	}
	if !found {
		t.Fatal("on-disk search did not find exact target")
	}
}
