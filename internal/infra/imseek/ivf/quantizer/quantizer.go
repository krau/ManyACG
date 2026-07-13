package quantizer

import (
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"

	"github.com/krau/ManyACG/internal/infra/imseek/hnsw"
)

type Quantizer interface {
	Search(x [][]byte, k int) []int64
	Centroids() [][]byte
	NList() int
	Save(path string) error
}

type HNSWQuantizer struct {
	index *hnsw.HNSW
}

func New(centroids [][]byte) *HNSWQuantizer {
	idx := hnsw.New(hnsw.DefaultParams())
	for _, c := range centroids {
		idx.Add(c)
	}
	return &HNSWQuantizer{index: idx}
}

func Open(path string) (*HNSWQuantizer, error) {
	idx, err := hnsw.Load(path)
	if err != nil {
		return nil, fmt.Errorf("open quantizer: %w", err)
	}
	return &HNSWQuantizer{index: idx}, nil
}

func (q *HNSWQuantizer) Search(x [][]byte, k int) []int64 {
	out := make([]int64, len(x)*k)
	if len(x) == 0 {
		return out
	}

	search := func(i int) {
		ids := q.index.SearchIDs(x[i], k)
		base := i * k
		for j := range k {
			if j < len(ids) {
				out[base+j] = ids[j]
			} else {
				out[base+j] = -1
			}
		}
	}

	workers := runtime.NumCPU()
	if len(x) < 2*workers {
		for i := range x {
			search(i)
		}
		return out
	}

	var next atomic.Int64
	var wg sync.WaitGroup
	wg.Add(workers)
	for range workers {
		go func() {
			defer wg.Done()
			for {
				i := int(next.Add(1)) - 1
				if i >= len(x) {
					return
				}
				search(i)
			}
		}()
	}
	wg.Wait()
	return out
}

func (q *HNSWQuantizer) Centroids() [][]byte { return q.index.Vectors() }

func (q *HNSWQuantizer) NList() int { return q.index.Len() }

func (q *HNSWQuantizer) Save(path string) error { return q.index.Save(path) }
