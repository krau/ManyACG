package hamming

import (
	"encoding/binary"
	"math/bits"
)

const CodeSize = 32

func Distance(a, b []byte) uint32 {
	n := len(a)
	if n != len(b) {
		panic("hamming: mismatched descriptor lengths")
	}
	var sum uint32
	i := 0
	for ; i+8 <= n; i += 8 {
		x := binary.LittleEndian.Uint64(a[i:]) ^ binary.LittleEndian.Uint64(b[i:])
		sum += uint32(bits.OnesCount64(x))
	}
	for ; i < n; i++ {
		sum += uint32(bits.OnesCount8(a[i] ^ b[i]))
	}
	return sum
}

type Neighbor struct {
	Index    int
	Distance uint32
}

func KNN(va []byte, vb [][]byte, k int) []Neighbor {
	if k <= 0 || len(vb) == 0 {
		return nil
	}
	if k > len(vb) {
		k = len(vb)
	}
	h := make([]Neighbor, 0, k)
	for i, v := range vb {
		d := Distance(va, v)
		if len(h) < k {
			h = append(h, Neighbor{Index: i, Distance: d})
			if len(h) == k {
				buildMaxHeap(h)
			}
			continue
		}
		if d < h[0].Distance {
			h[0] = Neighbor{Index: i, Distance: d}
			siftDown(h, 0)
		}
	}
	return h
}

func BatchKNN(va [][]byte, vb [][]byte, k int) [][]Neighbor {
	out := make([][]Neighbor, len(va))
	for i, q := range va {
		out[i] = KNN(q, vb, k)
	}
	return out
}

func buildMaxHeap(h []Neighbor) {
	for i := len(h)/2 - 1; i >= 0; i-- {
		siftDown(h, i)
	}
}

func siftDown(h []Neighbor, i int) {
	n := len(h)
	for {
		largest := i
		l, r := 2*i+1, 2*i+2
		if l < n && h[l].Distance > h[largest].Distance {
			largest = l
		}
		if r < n && h[r].Distance > h[largest].Distance {
			largest = r
		}
		if largest == i {
			return
		}
		h[i], h[largest] = h[largest], h[i]
		i = largest
	}
}
