package invlists

type InvertedLists interface {
	NList() int
	ListLen(listNo int) int
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
