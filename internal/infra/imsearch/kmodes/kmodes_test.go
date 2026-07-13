package kmodes

import (
	"math/rand"
	"testing"

	"github.com/krau/ManyACG/internal/infra/imsearch/hamming"
)

func makeClustered(rng *rand.Rand, centers [][]byte, perCenter, dim int) [][]byte {
	var data [][]byte
	for _, c := range centers {
		for range perCenter {
			v := make([]byte, dim)
			copy(v, c)
			// flip a few random bits (small noise)
			for range 3 {
				bit := rng.Intn(dim * 8)
				v[bit/8] ^= 1 << (bit % 8)
			}
			data = append(data, v)
		}
	}
	return data
}

func TestUpdateCentroidMajority(t *testing.T) {
	// 3 points; bit 0 set in 2 of 3 -> majority -> set. bit 1 set in 1 of 3 -> not.
	data := [][]byte{{0b01}, {0b01}, {0b10}}
	assignments := []int{0, 0, 0}
	centroids := make([][]byte, 1)
	frequency := make([]int, 1)
	updateCentroids(data, assignments, centroids, frequency, 1, 1)
	if frequency[0] != 3 {
		t.Fatalf("freq=%d want 3", frequency[0])
	}
	if centroids[0][0] != 0b01 {
		t.Fatalf("centroid=%08b want 00000001", centroids[0][0])
	}
}

func TestClusterRecoversCenters(t *testing.T) {
	rng := rand.New(rand.NewSource(99))
	dim := 8
	centers := [][]byte{
		{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF},
		{0xFF, 0x00, 0xFF, 0x00, 0xFF, 0x00, 0xFF, 0x00},
	}
	data := makeClustered(rng, centers, 50, dim)

	st := Cluster(data, 3, 20, InitKmeansPlusPlus, rng)
	if len(st.Centroids) != 3 {
		t.Fatalf("got %d centroids want 3", len(st.Centroids))
	}
	// Every seed center must be close to some produced centroid.
	for _, seed := range centers {
		best := uint32(1 << 30)
		for _, c := range st.Centroids {
			if d := hamming.Distance(seed, c); d < best {
				best = d
			}
		}
		if best > 4 {
			t.Fatalf("seed not recovered, min distance %d", best)
		}
	}
	// total frequency == number of points
	sum := 0
	for _, f := range st.Frequency {
		sum += f
	}
	if sum != len(data) {
		t.Fatalf("frequency sum %d want %d", sum, len(data))
	}
}

func TestClusterEmpty(t *testing.T) {
	st := Cluster(nil, 3, 10, InitRandom, nil)
	if len(st.Centroids) != 0 {
		t.Fatal("expected empty state")
	}
}

func TestImbalanceFactor(t *testing.T) {
	// perfectly balanced -> 1.0
	if got := ImbalanceFactor([]int{10, 10, 10}); got < 0.99 || got > 1.01 {
		t.Fatalf("balanced imbalance=%f want ~1.0", got)
	}
	// imbalanced -> > 1.0
	if got := ImbalanceFactor([]int{30, 0, 0}); got < 2.9 {
		t.Fatalf("imbalanced factor=%f want ~3.0", got)
	}
}

func TestIsqrt(t *testing.T) {
	cases := map[int]int{0: 0, 1: 1, 3: 1, 4: 2, 15: 3, 16: 4, 65536: 256, 65535: 255}
	for in, want := range cases {
		if got := isqrt(in); got != want {
			t.Errorf("isqrt(%d)=%d want %d", in, got, want)
		}
	}
}

func TestCluster2Level(t *testing.T) {
	rng := rand.New(rand.NewSource(5))
	dim := 8
	nc := 16
	// need n >= 30*nc = 480
	data := make([][]byte, 0, 600)
	for range 600 {
		v := make([]byte, dim)
		rng.Read(v)
		data = append(data, v)
	}
	st := Cluster2Level(data, nc, 10, InitRandom, rng)
	if len(st.Centroids) != nc {
		t.Fatalf("got %d centroids want %d", len(st.Centroids), nc)
	}
}
