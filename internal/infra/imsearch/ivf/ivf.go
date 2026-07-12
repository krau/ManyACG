package ivf

import (
	"runtime"
	"sync"
	"sync/atomic"

	"github.com/krau/ManyACG/internal/infra/imsearch/hamming"
	"github.com/krau/ManyACG/internal/infra/imsearch/ivf/invlists"
	"github.com/krau/ManyACG/internal/infra/imsearch/ivf/quantizer"
)

// Neighbor is a search hit: a vector id and its Hamming distance to the query.
type Neighbor struct {
	Distance uint32
	ID       uint64
}

type Index struct {
	quantizer quantizer.Quantizer
	invlists  invlists.InvertedLists
	threads   int
}

func New(q quantizer.Quantizer, il invlists.InvertedLists, threads int) *Index {
	if threads <= 0 {
		threads = runtime.NumCPU()
	}
	return &Index{quantizer: q, invlists: il, threads: threads}
}

// xor returns a XOR b into a fresh slice (a and b must be equal length).
func xor(a, b []byte) []byte {
	out := make([]byte, len(a))
	for i := range a {
		out[i] = a[i] ^ b[i]
	}
	return out
}

func (idx *Index) Add(data [][]byte, ids []uint64) {
	w, ok := idx.invlists.(invlists.Writable)
	if !ok {
		panic("ivf: Add requires writable inverted lists")
	}
	labels := idx.quantizer.Search(data, 1)
	centroids := idx.quantizer.Centroids()
	for i, vec := range data {
		listNo := labels[i]
		if listNo < 0 {
			continue
		}
		code := xor(vec, centroids[listNo])
		w.AddEntry(int(listNo), ids[i], code)
	}
}

func (idx *Index) Search(data [][]byte, k, nprobe int) []Neighbor {
	if len(data) == 0 {
		return nil
	}
	vlists := idx.quantizer.Search(data, nprobe)
	centroids := idx.quantizer.Centroids()

	// per-query bounded top-k, protected individually
	tops := make([]*topK, len(data))
	locks := make([]sync.Mutex, len(data))
	for i := range tops {
		tops[i] = newTopK(k)
	}

	// Group query descriptors by the list they probed, so each unique list is
	// read and decompressed exactly once (a hot list may be probed by many
	// descriptors). This avoids O(nqueries*nprobe) redundant decompressions.
	listQueries := make(map[int][]int)
	for i, listNo := range vlists {
		if listNo < 0 {
			continue
		}
		q := i / nprobe
		listQueries[int(listNo)] = append(listQueries[int(listNo)], q)
	}

	// Distribute unique lists across workers.
	uniqueLists := make([]int, 0, len(listQueries))
	for listNo := range listQueries {
		uniqueLists = append(uniqueLists, listNo)
	}

	var next atomic.Int64
	var wg sync.WaitGroup
	worker := func() {
		defer wg.Done()
		for {
			idxL := int(next.Add(1)) - 1
			if idxL >= len(uniqueLists) {
				return
			}
			listNo := uniqueLists[idxL]
			ids, codes, err := idx.invlists.GetList(listNo)
			if err != nil || len(ids) == 0 {
				continue
			}
			centroid := centroids[listNo]
			for _, q := range listQueries[listNo] {
				xq := xor(data[q], centroid)
				nn := hamming.KNN(xq, codes, k)
				locks[q].Lock()
				for _, n := range nn {
					tops[q].push(Neighbor{Distance: n.Distance, ID: ids[n.Index]})
				}
				locks[q].Unlock()
			}
		}
	}
	workers := min(idx.threads, len(uniqueLists))
	wg.Add(workers)
	for range workers {
		go worker()
	}
	wg.Wait()

	var out []Neighbor
	for _, t := range tops {
		out = append(out, t.items...)
	}
	return out
}

func (idx *Index) NList() int { return idx.invlists.NList() }
