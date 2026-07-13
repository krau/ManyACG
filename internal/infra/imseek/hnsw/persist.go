package hnsw

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

var magic = [8]byte{'I', 'M', 'H', 'N', 'S', 'W', '0', '1'}

func (h *HNSW) Save(path string) error {
	h.mu.RLock()
	defer h.mu.RUnlock()

	tmp := path + ".tmp"
	f, err := os.Create(tmp)
	if err != nil {
		return err
	}
	w := bufio.NewWriter(f)

	codeSize := 0
	if len(h.vectors) > 0 {
		codeSize = len(h.vectors[0])
	}

	writeU32 := func(v uint32) error { return binary.Write(w, binary.LittleEndian, v) }

	if _, err := w.Write(magic[:]); err != nil {
		return closeErr(f, tmp, err)
	}
	for _, v := range []uint32{
		uint32(h.params.M), uint32(h.params.M0), uint32(h.params.EfConstruction),
		uint32(h.params.EfSearch), uint32(h.params.MaxLayer),
	} {
		if err := writeU32(v); err != nil {
			return closeErr(f, tmp, err)
		}
	}
	if err := binary.Write(w, binary.LittleEndian, h.params.Seed); err != nil {
		return closeErr(f, tmp, err)
	}
	if err := writeU32(uint32(codeSize)); err != nil {
		return closeErr(f, tmp, err)
	}
	if err := writeU32(uint32(len(h.vectors))); err != nil {
		return closeErr(f, tmp, err)
	}
	if err := binary.Write(w, binary.LittleEndian, h.entryPoint); err != nil {
		return closeErr(f, tmp, err)
	}
	if err := binary.Write(w, binary.LittleEndian, int32(h.maxLevel)); err != nil {
		return closeErr(f, tmp, err)
	}
	for _, v := range h.vectors {
		if _, err := w.Write(v); err != nil {
			return closeErr(f, tmp, err)
		}
	}
	for _, nodeLayers := range h.links {
		if err := writeU32(uint32(len(nodeLayers))); err != nil {
			return closeErr(f, tmp, err)
		}
		for _, layer := range nodeLayers {
			if err := writeU32(uint32(len(layer))); err != nil {
				return closeErr(f, tmp, err)
			}
			for _, nb := range layer {
				if err := writeU32(nb); err != nil {
					return closeErr(f, tmp, err)
				}
			}
		}
	}
	if err := w.Flush(); err != nil {
		return closeErr(f, tmp, err)
	}
	if err := f.Close(); err != nil {
		os.Remove(tmp)
		return err
	}
	return os.Rename(tmp, path)
}

func closeErr(f *os.File, tmp string, err error) error {
	f.Close()
	os.Remove(tmp)
	return err
}

func Load(path string) (*HNSW, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	r := bufio.NewReader(f)

	var m [8]byte
	if _, err := io.ReadFull(r, m[:]); err != nil {
		return nil, err
	}
	if m != magic {
		return nil, fmt.Errorf("hnsw: bad magic %q", m)
	}

	readU32 := func() (uint32, error) {
		var v uint32
		err := binary.Read(r, binary.LittleEndian, &v)
		return v, err
	}

	vals := make([]uint32, 5)
	for i := range vals {
		if vals[i], err = readU32(); err != nil {
			return nil, err
		}
	}
	var seed int64
	if err := binary.Read(r, binary.LittleEndian, &seed); err != nil {
		return nil, err
	}
	p := Params{
		M: int(vals[0]), M0: int(vals[1]), EfConstruction: int(vals[2]),
		EfSearch: int(vals[3]), MaxLayer: int(vals[4]), Seed: seed,
	}
	h := New(p)

	codeSize, err := readU32()
	if err != nil {
		return nil, err
	}
	nVec, err := readU32()
	if err != nil {
		return nil, err
	}
	if err := binary.Read(r, binary.LittleEndian, &h.entryPoint); err != nil {
		return nil, err
	}
	var maxLevel int32
	if err := binary.Read(r, binary.LittleEndian, &maxLevel); err != nil {
		return nil, err
	}
	h.maxLevel = int(maxLevel)

	h.vectors = make([][]byte, nVec)
	for i := range h.vectors {
		buf := make([]byte, codeSize)
		if _, err := io.ReadFull(r, buf); err != nil {
			return nil, err
		}
		h.vectors[i] = buf
	}

	h.links = make([][][]uint32, nVec)
	for i := range h.links {
		nLayers, err := readU32()
		if err != nil {
			return nil, err
		}
		layers := make([][]uint32, nLayers)
		for l := range layers {
			nLinks, err := readU32()
			if err != nil {
				return nil, err
			}
			links := make([]uint32, nLinks)
			for j := range links {
				if links[j], err = readU32(); err != nil {
					return nil, err
				}
			}
			layers[l] = links
		}
		h.links[i] = layers
	}
	return h, nil
}
