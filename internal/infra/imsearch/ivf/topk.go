package ivf

type topK struct {
	items []Neighbor
	k     int
}

func newTopK(k int) *topK {
	return &topK{items: make([]Neighbor, 0, k), k: k}
}

func (t *topK) push(n Neighbor) {
	if t.k <= 0 {
		return
	}
	if len(t.items) < t.k {
		t.items = append(t.items, n)
		if len(t.items) == t.k {
			t.buildHeap()
		}
		return
	}
	if n.Distance < t.items[0].Distance {
		t.items[0] = n
		t.siftDown(0)
	}
}

func (t *topK) buildHeap() {
	for i := len(t.items)/2 - 1; i >= 0; i-- {
		t.siftDown(i)
	}
}

func (t *topK) siftDown(i int) {
	n := len(t.items)
	for {
		largest := i
		l, r := 2*i+1, 2*i+2
		if l < n && t.items[l].Distance > t.items[largest].Distance {
			largest = l
		}
		if r < n && t.items[r].Distance > t.items[largest].Distance {
			largest = r
		}
		if largest == i {
			return
		}
		t.items[i], t.items[largest] = t.items[largest], t.items[i]
		i = largest
	}
}
