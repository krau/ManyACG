package invlists

import "fmt"

type VStack struct {
	nlist int
	subs  []InvertedLists
}

func NewVStack(subs ...InvertedLists) (*VStack, error) {
	if len(subs) == 0 {
		return nil, fmt.Errorf("vstack: no inverted lists provided")
	}
	nlist := subs[0].NList()
	for _, s := range subs {
		if s.NList() != nlist {
			return nil, fmt.Errorf("vstack: mismatched nlist %d vs %d", s.NList(), nlist)
		}
	}
	return &VStack{nlist: nlist, subs: subs}, nil
}

func (v *VStack) NList() int { return v.nlist }

func (v *VStack) ListLen(listNo int) int {
	total := 0
	for _, s := range v.subs {
		total += s.ListLen(listNo)
	}
	return total
}

func (v *VStack) GetList(listNo int) ([]uint64, [][]byte, error) {
	var ids []uint64
	var codes [][]byte
	for _, s := range v.subs {
		sids, scodes, err := s.GetList(listNo)
		if err != nil {
			return nil, nil, err
		}
		ids = append(ids, sids...)
		codes = append(codes, scodes...)
	}
	return ids, codes, nil
}
