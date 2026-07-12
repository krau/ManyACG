package invlists

type ArrayInvertedLists struct {
	nlist int
	ids   [][]uint64
	codes [][][]byte
}

func NewArray(nlist int) *ArrayInvertedLists {
	return &ArrayInvertedLists{
		nlist: nlist,
		ids:   make([][]uint64, nlist),
		codes: make([][][]byte, nlist),
	}
}

func (a *ArrayInvertedLists) NList() int { return a.nlist }

func (a *ArrayInvertedLists) ListLen(listNo int) int { return len(a.ids[listNo]) }

func (a *ArrayInvertedLists) GetList(listNo int) ([]uint64, [][]byte, error) {
	return a.ids[listNo], a.codes[listNo], nil
}

func (a *ArrayInvertedLists) AddEntry(listNo int, id uint64, code []byte) {
	c := make([]byte, len(code))
	copy(c, code)
	a.ids[listNo] = append(a.ids[listNo], id)
	a.codes[listNo] = append(a.codes[listNo], c)
}
