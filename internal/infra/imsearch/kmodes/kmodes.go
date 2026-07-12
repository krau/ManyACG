package kmodes

import (
	"math"
	"math/rand"

	"github.com/krau/ManyACG/internal/infra/imsearch/hamming"
)

type InitMethod int

const (
	// InitRandom picks k distinct vectors uniformly at random.
	InitRandom InitMethod = iota
	// InitKmeansPlusPlus uses k-means++ style weighted seeding.
	InitKmeansPlusPlus
)

type State struct {
	// DistSum is the total Hamming distance from all points to their nearest
	// centroid at convergence.
	DistSum uint32
	// Centroids holds the cluster centers, each a code of CodeSize bytes.
	Centroids [][]byte
	// Frequency[i] is the number of points assigned to centroid i.
	Frequency []int
}

func codeSize(data [][]byte) int {
	if len(data) == 0 {
		return hamming.CodeSize
	}
	return len(data[0])
}

func cloneCode(c []byte) []byte {
	b := make([]byte, len(c))
	copy(b, c)
	return b
}

func initCentroids(data [][]byte, k int, method InitMethod, rng *rand.Rand) [][]byte {
	switch method {
	case InitKmeansPlusPlus:
		return initKmeansPlusPlus(data, k, rng)
	default:
		return initRandom(data, k, rng)
	}
}

func initRandom(data [][]byte, k int, rng *rand.Rand) [][]byte {
	perm := rng.Perm(len(data))
	if k > len(data) {
		k = len(data)
	}
	out := make([][]byte, k)
	for i := 0; i < k; i++ {
		out[i] = cloneCode(data[perm[i]])
	}
	return out
}

func initKmeansPlusPlus(data [][]byte, k int, rng *rand.Rand) [][]byte {
	centroids := make([][]byte, 0, k)
	centroids = append(centroids, cloneCode(data[rng.Intn(len(data))]))

	for len(centroids) < k {
		// For each point, the minimum distance to any existing centroid.
		weights := make([]float64, len(data))
		var total float64
		for i, x := range data {
			minD := uint32(math.MaxUint32)
			for _, c := range centroids {
				if d := hamming.Distance(x, c); d < minD {
					minD = d
				}
			}
			weights[i] = float64(minD)
			total += weights[i]
		}
		var next int
		if total == 0 {
			// All remaining points coincide with centroids; pick at random.
			next = rng.Intn(len(data))
		} else {
			next = weightedSample(weights, total, rng)
		}
		centroids = append(centroids, cloneCode(data[next]))
	}
	return centroids
}

func weightedSample(weights []float64, total float64, rng *rand.Rand) int {
	r := rng.Float64() * total
	var acc float64
	for i, w := range weights {
		acc += w
		if r < acc {
			return i
		}
	}
	return len(weights) - 1
}

func Cluster(data [][]byte, k, maxIter int, method InitMethod, rng *rand.Rand) State {
	if len(data) == 0 || k == 0 {
		return State{}
	}
	if rng == nil {
		rng = rand.New(rand.NewSource(rand.Int63()))
	}
	cs := codeSize(data)
	centroids := initCentroids(data, k, method, rng)

	var assignments []int
	distance := uint32(math.MaxUint32)
	frequency := make([]int, k)

	for range maxIter {
		newAssignments, newDistance := updateAssignments(data, centroids)
		if newDistance >= distance {
			break
		}
		assignments = newAssignments
		distance = newDistance

		for cid := range k {
			c, freq := updateCentroid(data, assignments, cid, cs)
			centroids[cid] = c
			frequency[cid] = freq
		}
	}

	return State{DistSum: distance, Centroids: centroids, Frequency: frequency}
}

func updateAssignments(data, centroids [][]byte) ([]int, uint32) {
	assignments := make([]int, len(data))
	var total uint32
	for i, point := range data {
		minD := uint32(math.MaxUint32)
		best := 0
		for j, c := range centroids {
			if d := hamming.Distance(point, c); d < minD {
				minD = d
				best = j
			}
		}
		assignments[i] = best
		total += minD
	}
	return assignments, total
}

func updateCentroid(data [][]byte, assignments []int, clusterID, cs int) ([]byte, int) {
	var points [][]byte
	for i, a := range assignments {
		if a == clusterID {
			points = append(points, data[i])
		}
	}
	if len(points) == 0 {
		return make([]byte, cs), 0
	}

	centroid := make([]byte, cs)
	half := uint32(len(points)) / 2
	for bytePos := range cs {
		var bitCounts [8]uint32
		for _, p := range points {
			bv := p[bytePos]
			for bit := range 8 {
				if (bv>>bit)&1 == 1 {
					bitCounts[bit]++
				}
			}
		}
		var nb byte
		for bit := range 8 {
			if bitCounts[bit] > half {
				nb |= 1 << bit
			}
		}
		centroid[bytePos] = nb
	}
	return centroid, len(points)
}

func ImbalanceFactor(hist []int) float32 {
	var tot, uf float64
	for _, h := range hist {
		hf := float64(h)
		tot += hf
		uf += hf * hf
	}
	if tot == 0 {
		return 0
	}
	return float32(uf * float64(len(hist)) / (tot * tot))
}

func isqrt(n int) int {
	if n < 0 {
		return 0
	}
	x := int(math.Sqrt(float64(n)))
	for (x+1)*(x+1) <= n {
		x++
	}
	for x*x > n {
		x--
	}
	return x
}

func Cluster2Level(x [][]byte, nc, maxIter int, method InitMethod, rng *rand.Rand) State {
	n := len(x)
	if n < 30*nc {
		panic("kmodes: vector count must be >= 30 * nc")
	}
	if rng == nil {
		rng = rand.New(rand.NewSource(rand.Int63()))
	}
	nc1 := max(isqrt(nc), 1)

	// Level 1: train on a subset.
	n1 := min(nc1*1024, n)
	ks := Cluster(x[:n1], nc1, maxIter, method, rng)

	// Assign ALL vectors to level-1 centroids.
	assignments, _ := updateAssignments(x, ks.Centroids)
	xc := make([][][]byte, nc1)
	for i, a := range assignments {
		xc[a] = append(xc[a], x[i])
	}

	// Distribute nc sub-centroids across buckets weighted by bucket size, using
	// prefix-sum + reverse-difference so the totals sum exactly to nc.
	bcSum := make([]int, nc1)
	acc := 0
	for i := range xc {
		acc += len(xc[i])
		bcSum[i] = acc
	}
	nc2 := make([]int, nc1)
	last := bcSum[nc1-1]
	for i := range nc2 {
		if last == 0 {
			nc2[i] = 0
		} else {
			nc2[i] = bcSum[i] * nc / last
		}
	}
	for i := nc1 - 1; i >= 1; i-- {
		nc2[i] -= nc2[i-1]
	}

	var final State
	for i := range nc1 {
		if nc2[i] > 0 {
			sub := Cluster(xc[i], nc2[i], maxIter, method, rng)
			final.DistSum += sub.DistSum
			final.Centroids = append(final.Centroids, sub.Centroids...)
			final.Frequency = append(final.Frequency, sub.Frequency...)
		}
	}
	if len(final.Centroids) != nc {
		panic("kmodes: 2-level produced wrong centroid count")
	}
	return final
}
