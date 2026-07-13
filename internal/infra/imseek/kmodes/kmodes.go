package kmodes

import (
	"math"
	"math/rand"
	"runtime"
	"sync"

	"github.com/krau/ManyACG/internal/infra/imseek/hamming"
)

type InitMethod int

const (
	InitRandom InitMethod = iota
	InitKmeansPlusPlus
)

type State struct {
	DistSum   uint32
	Centroids [][]byte
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

	assignments := make([]int, len(data))
	distance := uint32(math.MaxUint32)
	frequency := make([]int, k)

	for range maxIter {
		newDistance := updateAssignments(data, centroids, assignments)
		if newDistance >= distance {
			break
		}
		distance = newDistance

		updateCentroids(data, assignments, centroids, frequency, cs, k)
	}

	return State{DistSum: distance, Centroids: centroids, Frequency: frequency}
}

func updateAssignments(data, centroids [][]byte, assignments []int) uint32 {
	n := len(data)
	nc := len(centroids)
	if n == 0 || nc == 0 {
		return 0
	}

	workers := max(min(runtime.NumCPU(), n), 1)

	type partial struct {
		total uint32
	}

	results := make([]partial, workers)
	chunk := (n + workers - 1) / workers

	var wg sync.WaitGroup
	for w := range workers {
		lo := w * chunk
		hi := min(lo+chunk, n)
		if lo >= hi {
			continue
		}
		wg.Add(1)
		go func(lo, hi, wid int) {
			defer wg.Done()
			var sub uint32
			for i := lo; i < hi; i++ {
				point := data[i]
				minD := uint32(math.MaxUint32)
				best := 0
				for j := range nc {
					if d := hamming.Distance(point, centroids[j]); d < minD {
						minD = d
						best = j
					}
				}
				assignments[i] = best
				sub += minD
			}
			results[wid].total = sub
		}(lo, hi, w)
	}
	wg.Wait()

	var total uint32
	for _, r := range results {
		total += r.total
	}
	return total
}

func updateCentroids(data [][]byte, assignments []int, centroids [][]byte, frequency []int, cs, k int) {
	nBits := k * cs * 8
	counts := make([]uint32, k)

	n := len(data)
	workers := max(min(runtime.NumCPU(), n), 1)

	type workerAcc struct {
		bits   []uint32
		counts []uint32
	}
	accs := make([]workerAcc, workers)
	for i := range workers {
		accs[i] = workerAcc{
			bits:   make([]uint32, nBits),
			counts: make([]uint32, k),
		}
	}

	chunk := (n + workers - 1) / workers
	var wg sync.WaitGroup
	for w := 0; w < workers; w++ {
		lo := w * chunk
		hi := min(lo+chunk, n)
		if lo >= hi {
			continue
		}
		wg.Add(1)
		go func(lo, hi, wid int) {
			defer wg.Done()
			lb := accs[wid].bits
			lc := accs[wid].counts
			for i := lo; i < hi; i++ {
				cid := assignments[i]
				lc[cid]++
				p := data[i]
				for bp := range cs {
					bv := p[bp]
					base := cid*cs*8 + bp*8
					for bit := range 8 {
						if (bv>>bit)&1 == 1 {
							lb[base+bit]++
						}
					}
				}
			}
		}(lo, hi, w)
	}
	wg.Wait()

	bitCounts := make([]uint32, nBits)
	for w := range accs {
		for i := range nBits {
			bitCounts[i] += accs[w].bits[i]
		}
		for cid := range k {
			counts[cid] += accs[w].counts[cid]
		}
	}

	for cid := range k {
		frequency[cid] = int(counts[cid])
		if counts[cid] == 0 {
			centroids[cid] = make([]byte, cs)
			continue
		}
		half := counts[cid] / 2
		centroid := make([]byte, cs)
		for bp := range cs {
			base := cid*cs*8 + bp*8
			var nb byte
			for bit := range 8 {
				if bitCounts[base+bit] > half {
					nb |= 1 << bit
				}
			}
			centroid[bp] = nb
		}
		centroids[cid] = centroid
	}
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

	n1 := min(nc1*1024, n)
	ks := Cluster(x[:n1], nc1, maxIter, method, rng)

	assignments := make([]int, len(x))
	updateAssignments(x, ks.Centroids, assignments)
	xc := make([][][]byte, nc1)
	for i, a := range assignments {
		xc[a] = append(xc[a], x[i])
	}

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
