package hnsw

import (
	"math"
	"math/rand"
	"sync"

	"github.com/krau/ManyACG/internal/infra/imseek/hamming"
)

type Params struct {
	M                    int // max neighbors per node on layers > 0
	M0                   int // max neighbors on layer 0 (typically 2*M)
	EfConstruction       int // candidate list size during insertion
	EfSearch             int // candidate list size during search
	MaxLayer             int // hard cap on layer count
	Seed                 int64
	SimpleNeighborSelect bool
}

func DefaultParams() Params {
	return Params{
		M:              32,
		M0:             64,
		EfConstruction: 128,
		EfSearch:       16,
		MaxLayer:       16,
		Seed:           0x5eed,
	}
}

type HNSW struct {
	params   Params
	mu       sync.RWMutex
	rng      *rand.Rand
	levelMul float64

	vectors    [][]byte
	links      [][][]uint32
	entryPoint int32
	maxLevel   int
}

func New(p Params) *HNSW {
	if p.M0 == 0 {
		p.M0 = 2 * p.M
	}
	return &HNSW{
		params:     p,
		rng:        rand.New(rand.NewSource(p.Seed)),
		levelMul:   1.0 / math.Log(float64(p.M)),
		entryPoint: -1,
		maxLevel:   -1,
	}
}

func (h *HNSW) Len() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.vectors)
}

func (h *HNSW) Vectors() [][]byte {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.vectors
}

func (h *HNSW) dist(a, b []byte) uint32 { return hamming.Distance(a, b) }

func (h *HNSW) randomLevel() int {
	r := h.rng.Float64()
	if r <= 0 {
		r = math.SmallestNonzeroFloat64
	}
	lvl := min(int(-math.Log(r)*h.levelMul), h.params.MaxLayer)
	return lvl
}

func (h *HNSW) Add(vec []byte) uint32 {
	h.mu.Lock()
	defer h.mu.Unlock()

	v := make([]byte, len(vec))
	copy(v, vec)
	id := uint32(len(h.vectors))
	h.vectors = append(h.vectors, v)

	level := h.randomLevel()
	nodeLinks := make([][]uint32, level+1)
	h.links = append(h.links, nodeLinks)

	if h.entryPoint == -1 {
		h.entryPoint = int32(id)
		h.maxLevel = level
		return id
	}

	curr := uint32(h.entryPoint)
	currDist := h.dist(v, h.vectors[curr])

	for lc := h.maxLevel; lc > level; lc-- {
		curr, currDist = h.greedyClosest(v, curr, currDist, lc)
	}

	start := min(level, h.maxLevel)
	for lc := start; lc >= 0; lc-- {
		candidates := h.searchLayer(v, curr, h.params.EfConstruction, lc)
		m := h.params.M
		if lc == 0 {
			m = h.params.M0
		}
		selected := h.selectNeighbors(candidates, m)
		h.links[id][lc] = selected
		for _, nb := range selected {
			h.links[nb][lc] = append(h.links[nb][lc], id)
			if len(h.links[nb][lc]) > m {
				h.links[nb][lc] = h.pruneNeighbors(nb, lc, m)
			}
		}
		if len(candidates) > 0 {
			curr = candidates[0].id
		}
	}

	if level > h.maxLevel {
		h.maxLevel = level
		h.entryPoint = int32(id)
	}
	return id
}

func (h *HNSW) greedyClosest(v []byte, curr uint32, currDist uint32, layer int) (uint32, uint32) {
	for {
		improved := false
		if layer < len(h.links[curr]) {
			for _, nb := range h.links[curr][layer] {
				d := h.dist(v, h.vectors[nb])
				if d < currDist {
					currDist = d
					curr = nb
					improved = true
				}
			}
		}
		if !improved {
			return curr, currDist
		}
	}
}

type candidate struct {
	id   uint32
	dist uint32
}

type searchWorkspace struct {
	visited map[uint32]struct{}
	cand    minHeap
	res     maxHeap
}

var searchPool = sync.Pool{
	New: func() any {
		return &searchWorkspace{visited: make(map[uint32]struct{}, 256)}
	},
}

func (h *HNSW) searchLayer(v []byte, ep uint32, ef, layer int) []candidate {
	ws := searchPool.Get().(*searchWorkspace)
	visited := ws.visited
	cand := &ws.cand
	res := &ws.res
	defer func() {
		clear(visited)
		*cand = (*cand)[:0]
		*res = (*res)[:0]
		searchPool.Put(ws)
	}()

	d0 := h.dist(v, h.vectors[ep])
	visited[ep] = struct{}{}

	*cand = append(*cand, candidate{ep, d0})
	*res = append(*res, candidate{ep, d0})

	for len(*cand) > 0 {
		c := cand.popMin()
		if len(*res) >= ef && c.dist > (*res)[0].dist {
			break
		}
		if layer < len(h.links[c.id]) {
			for _, nb := range h.links[c.id][layer] {
				if _, ok := visited[nb]; ok {
					continue
				}
				visited[nb] = struct{}{}
				d := h.dist(v, h.vectors[nb])
				if len(*res) < ef || d < (*res)[0].dist {
					cand.pushMin(candidate{nb, d})
					res.pushMax(candidate{nb, d})
					if len(*res) > ef {
						res.popMax()
					}
				}
			}
		}
	}

	out := make([]candidate, len(*res))
	for i := len(out) - 1; i >= 0; i-- {
		out[i] = res.popMax()
	}
	return out
}

func (h *HNSW) selectNeighbors(candidates []candidate, m int) []uint32 {
	if h.params.SimpleNeighborSelect {
		if len(candidates) > m {
			candidates = candidates[:m]
		}
		out := make([]uint32, len(candidates))
		for i, c := range candidates {
			out[i] = c.id
		}
		return out
	}
	if len(candidates) <= m {
		out := make([]uint32, len(candidates))
		for i, c := range candidates {
			out[i] = c.id
		}
		return out
	}

	selected := make([]uint32, 0, m)
	used := make([]bool, len(candidates))
	for i, c := range candidates {
		if len(selected) >= m {
			break
		}
		good := true
		for _, sid := range selected {
			if h.dist(h.vectors[c.id], h.vectors[sid]) < c.dist {
				good = false
				break
			}
		}
		if good {
			selected = append(selected, c.id)
			used[i] = true
		}
	}
	for i, c := range candidates {
		if len(selected) >= m {
			break
		}
		if !used[i] {
			selected = append(selected, c.id)
		}
	}
	return selected
}

func (h *HNSW) pruneNeighbors(node uint32, layer, m int) []uint32 {
	nbs := h.links[node][layer]
	if len(nbs) <= m {
		return nbs
	}
	nodeVec := h.vectors[node]
	cands := make([]candidate, len(nbs))
	for i, nb := range nbs {
		cands[i] = candidate{nb, h.dist(nodeVec, h.vectors[nb])}
	}
	sortCandidates(cands)
	return h.selectNeighbors(cands, m)
}

type Result struct {
	ID       uint32
	Distance uint32
}

func (h *HNSW) Search(query []byte, k int) []Result {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if h.entryPoint == -1 || k <= 0 {
		return nil
	}
	curr := uint32(h.entryPoint)
	currDist := h.dist(query, h.vectors[curr])
	for lc := h.maxLevel; lc > 0; lc-- {
		curr, currDist = h.greedyClosest(query, curr, currDist, lc)
	}
	ef := max(h.params.EfSearch, k)
	res := h.searchLayer(query, curr, ef, 0)
	if len(res) > k {
		res = res[:k]
	}
	out := make([]Result, len(res))
	for i, c := range res {
		out[i] = Result{ID: c.id, Distance: c.dist}
	}
	return out
}

func (h *HNSW) SearchIDs(query []byte, k int) []int64 {
	res := h.Search(query, k)
	out := make([]int64, len(res))
	for i, c := range res {
		out[i] = int64(c.ID)
	}
	return out
}

func sortCandidates(c []candidate) {
	for i := 1; i < len(c); i++ {
		for j := i; j > 0 && c[j].dist < c[j-1].dist; j-- {
			c[j], c[j-1] = c[j-1], c[j]
		}
	}
}

type minHeap []candidate

func (h *minHeap) pushMin(c candidate) {
	*h = append(*h, c)
	n := len(*h) - 1
	for n > 0 {
		p := (n - 1) / 2
		if (*h)[p].dist <= (*h)[n].dist {
			break
		}
		(*h)[p], (*h)[n] = (*h)[n], (*h)[p]
		n = p
	}
}

func (h *minHeap) popMin() candidate {
	old := *h
	n := len(old)
	top := old[0]
	old[0] = old[n-1]
	*h = old[:n-1]
	h.siftDownMin(0)
	return top
}

func (h *minHeap) siftDownMin(i int) {
	n := len(*h)
	for {
		l := 2*i + 1
		if l >= n {
			return
		}
		smallest := l
		if r := l + 1; r < n && (*h)[r].dist < (*h)[l].dist {
			smallest = r
		}
		if (*h)[smallest].dist >= (*h)[i].dist {
			return
		}
		(*h)[i], (*h)[smallest] = (*h)[smallest], (*h)[i]
		i = smallest
	}
}

type maxHeap []candidate

func (h *maxHeap) pushMax(c candidate) {
	*h = append(*h, c)
	n := len(*h) - 1
	for n > 0 {
		p := (n - 1) / 2
		if (*h)[p].dist >= (*h)[n].dist {
			break
		}
		(*h)[p], (*h)[n] = (*h)[n], (*h)[p]
		n = p
	}
}

func (h *maxHeap) popMax() candidate {
	old := *h
	n := len(old)
	top := old[0]
	old[0] = old[n-1]
	*h = old[:n-1]
	h.siftDownMax(0)
	return top
}

func (h *maxHeap) siftDownMax(i int) {
	n := len(*h)
	for {
		l := 2*i + 1
		if l >= n {
			return
		}
		largest := l
		if r := l + 1; r < n && (*h)[r].dist > (*h)[l].dist {
			largest = r
		}
		if (*h)[largest].dist <= (*h)[i].dist {
			return
		}
		(*h)[i], (*h)[largest] = (*h)[largest], (*h)[i]
		i = largest
	}
}
