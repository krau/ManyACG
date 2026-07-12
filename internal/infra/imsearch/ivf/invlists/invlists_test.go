package invlists

import (
	"math/rand"
	"path/filepath"
	"testing"
)

func randCodes(rng *rand.Rand, n, cs int) ([]uint64, [][]byte) {
	ids := make([]uint64, n)
	codes := make([][]byte, n)
	for i := range n {
		ids[i] = uint64(rng.Int63())
		codes[i] = make([]byte, cs)
		rng.Read(codes[i])
	}
	return ids, codes
}

func TestArrayInvertedLists(t *testing.T) {
	a := NewArray(4)
	if a.NList() != 4 {
		t.Fatalf("nlist=%d", a.NList())
	}
	a.AddEntry(1, 100, []byte{1, 2, 3, 4, 5, 6, 7, 8})
	a.AddEntry(1, 200, []byte{8, 7, 6, 5, 4, 3, 2, 1})
	if a.ListLen(1) != 2 {
		t.Fatalf("listlen=%d", a.ListLen(1))
	}
	ids, codes, _ := a.GetList(1)
	if ids[0] != 100 || ids[1] != 200 {
		t.Fatalf("ids=%v", ids)
	}
	if codes[0][0] != 1 || codes[1][0] != 8 {
		t.Fatalf("codes wrong")
	}
}

func TestAddEntryCopiesCode(t *testing.T) {
	a := NewArray(1)
	code := []byte{1, 2, 3, 4, 5, 6, 7, 8}
	a.AddEntry(0, 1, code)
	code[0] = 99
	_, codes, _ := a.GetList(0)
	if codes[0][0] != 1 {
		t.Fatal("AddEntry did not copy code")
	}
}

func TestOnDiskRoundTrip(t *testing.T) {
	rng := rand.New(rand.NewSource(11))
	const nlist, cs = 20, 8
	a := NewArray(nlist)
	want := make(map[int]struct {
		ids   []uint64
		codes [][]byte
	})
	for i := range nlist {
		// vary list sizes, leave some empty
		n := rng.Intn(50)
		if i%7 == 0 {
			n = 0
		}
		ids, codes := randCodes(rng, n, cs)
		for j := 0; j < n; j++ {
			a.AddEntry(i, ids[j], codes[j])
		}
		want[i] = struct {
			ids   []uint64
			codes [][]byte
		}{ids, codes}
	}

	path := filepath.Join(t.TempDir(), "invlists.bin")
	if err := Save(a, path, cs, 3); err != nil {
		t.Fatal(err)
	}
	od, err := LoadOnDisk(path)
	if err != nil {
		t.Fatal(err)
	}
	defer od.Close()

	if od.NList() != nlist {
		t.Fatalf("nlist=%d want %d", od.NList(), nlist)
	}
	if od.CodeSize() != cs {
		t.Fatalf("codesize=%d want %d", od.CodeSize(), cs)
	}
	for i := range nlist {
		gotIDs, gotCodes, err := od.GetList(i)
		if err != nil {
			t.Fatal(err)
		}
		w := want[i]
		if len(gotIDs) != len(w.ids) {
			t.Fatalf("list %d len %d want %d", i, len(gotIDs), len(w.ids))
		}
		for j := range w.ids {
			if gotIDs[j] != w.ids[j] {
				t.Fatalf("list %d id %d = %d want %d", i, j, gotIDs[j], w.ids[j])
			}
			for b := range w.codes[j] {
				if gotCodes[j][b] != w.codes[j][b] {
					t.Fatalf("list %d code %d byte %d mismatch", i, j, b)
				}
			}
		}
	}
}

func TestVStackMerge(t *testing.T) {
	a := NewArray(3)
	b := NewArray(3)
	a.AddEntry(0, 1, []byte{1, 0, 0, 0, 0, 0, 0, 0})
	a.AddEntry(1, 2, []byte{2, 0, 0, 0, 0, 0, 0, 0})
	b.AddEntry(0, 3, []byte{3, 0, 0, 0, 0, 0, 0, 0})
	b.AddEntry(2, 4, []byte{4, 0, 0, 0, 0, 0, 0, 0})

	v, err := NewVStack(a, b)
	if err != nil {
		t.Fatal(err)
	}
	if v.ListLen(0) != 2 {
		t.Fatalf("list0 len=%d want 2", v.ListLen(0))
	}
	ids, _, _ := v.GetList(0)
	if len(ids) != 2 || ids[0] != 1 || ids[1] != 3 {
		t.Fatalf("merged ids=%v", ids)
	}
}

func TestVStackMismatchedNList(t *testing.T) {
	if _, err := NewVStack(NewArray(3), NewArray(4)); err == nil {
		t.Fatal("expected error for mismatched nlist")
	}
}
