package invlists

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/klauspost/compress/zstd"
)

const headerFixed = 16 // nlist u64 + codeSize u64

// Save writes an InvertedLists to path atomically using zstd compression.
func Save(il InvertedLists, path string, codeSize, level int) error {
	tmp := path + ".tmp"
	f, err := os.Create(tmp)
	if err != nil {
		return err
	}
	cleanup := func(e error) error {
		f.Close()
		os.Remove(tmp)
		return e
	}

	nlist := il.NList()
	listLen := make([]uint64, nlist)
	listOffset := make([]uint64, nlist)
	listSize := make([]uint64, nlist)
	listSplit := make([]uint64, nlist)

	enc, err := zstd.NewWriter(nil, zstd.WithEncoderLevel(zstd.EncoderLevelFromZstd(level)))
	if err != nil {
		return cleanup(err)
	}
	defer enc.Close()

	w := bufio.NewWriter(f)

	// Reserve header space.
	headerSize := headerFixed + 4*8*nlist
	if _, err := f.Seek(int64(headerSize), io.SeekStart); err != nil {
		return cleanup(err)
	}

	offset := uint64(headerSize)
	idBuf := make([]byte, 0)
	for i := range nlist {
		ids, codes, err := il.GetList(i)
		if err != nil {
			return cleanup(err)
		}
		if len(ids) == 0 {
			continue
		}
		// serialize ids as u64 LE
		if cap(idBuf) < len(ids)*8 {
			idBuf = make([]byte, len(ids)*8)
		}
		idBuf = idBuf[:len(ids)*8]
		for j, id := range ids {
			binary.LittleEndian.PutUint64(idBuf[j*8:], id)
		}
		// flatten codes
		flat := make([]byte, 0, len(codes)*codeSize)
		for _, c := range codes {
			flat = append(flat, c...)
		}

		compIDs := enc.EncodeAll(idBuf, nil)
		compCodes := enc.EncodeAll(flat, nil)

		if _, err := w.Write(compIDs); err != nil {
			return cleanup(err)
		}
		if _, err := w.Write(compCodes); err != nil {
			return cleanup(err)
		}

		listLen[i] = uint64(len(ids))
		listOffset[i] = offset
		listSize[i] = uint64(len(compIDs) + len(compCodes))
		listSplit[i] = uint64(len(compIDs))
		offset += listSize[i]
	}
	if err := w.Flush(); err != nil {
		return cleanup(err)
	}

	// Write header at the start.
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		return cleanup(err)
	}
	hw := bufio.NewWriter(f)
	if err := binary.Write(hw, binary.LittleEndian, uint64(nlist)); err != nil {
		return cleanup(err)
	}
	if err := binary.Write(hw, binary.LittleEndian, uint64(codeSize)); err != nil {
		return cleanup(err)
	}
	for _, arr := range [][]uint64{listLen, listOffset, listSize, listSplit} {
		if err := binary.Write(hw, binary.LittleEndian, arr); err != nil {
			return cleanup(err)
		}
	}
	if err := hw.Flush(); err != nil {
		return cleanup(err)
	}
	if err := f.Close(); err != nil {
		os.Remove(tmp)
		return err
	}
	return os.Rename(tmp, path)
}

type OnDisk struct {
	nlist      int
	codeSize   int
	listLen    []uint64
	listOffset []uint64
	listSize   []uint64
	listSplit  []uint64
	file       *os.File
	dec        *zstd.Decoder

	cacheMu sync.Mutex
	cache   map[int]*listCache
	maxCache int
}

type listCache struct {
	ids   []uint64
	codes [][]byte
}

func LoadOnDisk(path string) (*OnDisk, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	r := bufio.NewReader(f)
	var nlist, codeSize uint64
	if err := binary.Read(r, binary.LittleEndian, &nlist); err != nil {
		f.Close()
		return nil, err
	}
	if err := binary.Read(r, binary.LittleEndian, &codeSize); err != nil {
		f.Close()
		return nil, err
	}
	dec, err := zstd.NewReader(nil)
	if err != nil {
		f.Close()
		return nil, err
	}
	od := &OnDisk{
		nlist:      int(nlist),
		codeSize:   int(codeSize),
		listLen:    make([]uint64, nlist),
		listOffset: make([]uint64, nlist),
		listSize:   make([]uint64, nlist),
		listSplit:  make([]uint64, nlist),
		file:       f,
		dec:        dec,
		cache:      make(map[int]*listCache),
		maxCache:   512,
	}
	for _, arr := range [][]uint64{od.listLen, od.listOffset, od.listSize, od.listSplit} {
		if err := binary.Read(r, binary.LittleEndian, arr); err != nil {
			f.Close()
			return nil, err
		}
	}
	return od, nil
}

func (o *OnDisk) CodeSize() int { return o.codeSize }

func (o *OnDisk) NList() int { return o.nlist }

func (o *OnDisk) ListLen(listNo int) int { return int(o.listLen[listNo]) }

func (o *OnDisk) GetList(listNo int) ([]uint64, [][]byte, error) {
	n := int(o.listLen[listNo])
	if n == 0 {
		return nil, nil, nil
	}

	o.cacheMu.Lock()
	if c, ok := o.cache[listNo]; ok {
		o.cacheMu.Unlock()
		return c.ids, c.codes, nil
	}
	o.cacheMu.Unlock()

	size := int(o.listSize[listNo])
	offset := int64(o.listOffset[listNo])
	split := int(o.listSplit[listNo])

	buf := make([]byte, size)
	if _, err := o.file.ReadAt(buf, offset); err != nil {
		return nil, nil, fmt.Errorf("read list %d: %w", listNo, err)
	}

	idBytes, err := o.dec.DecodeAll(buf[:split], make([]byte, 0, n*8))
	if err != nil {
		return nil, nil, fmt.Errorf("decompress ids list %d: %w", listNo, err)
	}
	codeBytes, err := o.dec.DecodeAll(buf[split:], make([]byte, 0, n*o.codeSize))
	if err != nil {
		return nil, nil, fmt.Errorf("decompress codes list %d: %w", listNo, err)
	}

	ids := make([]uint64, n)
	for i := range n {
		ids[i] = binary.LittleEndian.Uint64(idBytes[i*8:])
	}
	codes := make([][]byte, n)
	for i := range n {
		codes[i] = codeBytes[i*o.codeSize : (i+1)*o.codeSize]
	}

	o.cacheMu.Lock()
	if len(o.cache) >= o.maxCache {
		for k := range o.cache {
			delete(o.cache, k)
			break
		}
	}
	o.cache[listNo] = &listCache{ids: ids, codes: codes}
	o.cacheMu.Unlock()

	return ids, codes, nil
}

func (o *OnDisk) Close() error {
	if o.dec != nil {
		o.dec.Close()
		o.dec = nil
	}
	return o.file.Close()
}
