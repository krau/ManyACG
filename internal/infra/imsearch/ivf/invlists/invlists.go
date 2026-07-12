package invlists

type InvertedLists interface {
	// NList returns the number of lists.
	NList() int
	// ListLen returns the number of entries in a list.
	ListLen(listNo int) int
	// GetList returns the ids and codes stored in a list. The returned slices
	// must not be mutated by the caller.
	GetList(listNo int) (ids []uint64, codes [][]byte, err error)
}

type Writable interface {
	InvertedLists
	AddEntry(listNo int, id uint64, code []byte)
}

func imbalanceFactor(hist []int) float32 {
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

func Imbalance(il InvertedLists) float32 {
	hist := make([]int, il.NList())
	for i := range hist {
		hist[i] = il.ListLen(i)
	}
	return imbalanceFactor(hist)
}
