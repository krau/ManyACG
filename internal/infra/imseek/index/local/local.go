package local

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/krau/ManyACG/internal/infra/imseek/config"
	"github.com/krau/ManyACG/internal/infra/imseek/index"
	"github.com/krau/ManyACG/internal/infra/imseek/ivf"
	"github.com/krau/ManyACG/internal/infra/imseek/ivf/invlists"
	"github.com/krau/ManyACG/internal/infra/imseek/ivf/quantizer"
	"github.com/krau/ManyACG/internal/infra/imseek/kmodes"
)

type Backend struct {
	confDir  string
	codeSize int
}

func New(confDir string, codeSize int) *Backend {
	return &Backend{confDir: confDir, codeSize: codeSize}
}

func (b *Backend) quantizerPath() string {
	return filepath.Join(b.confDir, config.QuantizerFile)
}

func (b *Backend) invlistsPath() string {
	return filepath.Join(b.confDir, config.InvlistsFile)
}

func (b *Backend) subIndexPath(i int) string {
	return filepath.Join(b.confDir, fmt.Sprintf("invlists.%d", i))
}

func (b *Backend) countShards() int {
	n := 0
	for {
		if _, err := os.Stat(b.subIndexPath(n)); os.IsNotExist(err) {
			break
		}
		n++
	}
	return n
}

func (b *Backend) CountShards() int { return b.countShards() }

func (b *Backend) Compact() error {
	nShards := b.countShards()
	if nShards == 0 {
		return nil
	}

	var subs []invlists.InvertedLists
	var closers []*invlists.OnDisk
	defer func() {
		for _, c := range closers {
			c.Close()
		}
	}()

	if _, err := os.Stat(b.invlistsPath()); err == nil {
		od, err := invlists.LoadOnDisk(b.invlistsPath())
		if err != nil {
			return fmt.Errorf("compact: load invlists: %w", err)
		}
		subs = append(subs, od)
		closers = append(closers, od)
	}
	for i := range nShards {
		od, err := invlists.LoadOnDisk(b.subIndexPath(i))
		if err != nil {
			return fmt.Errorf("compact: load shard %d: %w", i, err)
		}
		subs = append(subs, od)
		closers = append(closers, od)
	}

	v, err := invlists.NewVStack(subs...)
	if err != nil {
		return fmt.Errorf("compact: vstack: %w", err)
	}
	tmpPath := b.invlistsPath() + ".compact"
	if err := invlists.Save(v, tmpPath, b.codeSize, 3); err != nil {
		return fmt.Errorf("compact: save: %w", err)
	}
	for _, c := range closers {
		c.Close()
	}
	closers = nil

	if err := os.Rename(tmpPath, b.invlistsPath()); err != nil {
		return fmt.Errorf("compact: rename: %w", err)
	}
	for i := range nShards {
		os.Remove(b.subIndexPath(i))
	}
	return nil
}

func (b *Backend) Trained(_ context.Context) bool {
	_, err := os.Stat(b.quantizerPath())
	return err == nil
}

func (b *Backend) DeleteVectors(_ context.Context, lo, hi uint64) error {
	if lo > hi {
		return nil
	}
	if _, err := os.Stat(b.invlistsPath()); err == nil {
		if err := filterInvlistsFile(b.invlistsPath(), b.codeSize, lo, hi); err != nil {
			return fmt.Errorf("local delete main invlists: %w", err)
		}
	}
	for i := 0; ; i++ {
		path := b.subIndexPath(i)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			break
		}
		if err := filterInvlistsFile(path, b.codeSize, lo, hi); err != nil {
			return fmt.Errorf("local delete shard %d: %w", i, err)
		}
	}
	return nil
}

func filterInvlistsFile(path string, codeSize int, lo, hi uint64) error {
	od, err := invlists.LoadOnDisk(path)
	if err != nil {
		return err
	}
	defer od.Close()

	arr := invlists.NewArray(od.NList())
	for listNo := 0; listNo < od.NList(); listNo++ {
		ids, codes, err := od.GetList(listNo)
		if err != nil {
			return err
		}
		for j, id := range ids {
			if id >= lo && id <= hi {
				continue
			}
			arr.AddEntry(listNo, id, codes[j])
		}
	}
	tmp := path + ".del"
	if err := invlists.Save(arr, tmp, codeSize, 3); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func (b *Backend) Train(_ context.Context, samples [][]byte, opts index.TrainOptions) (index.TrainResult, error) {
	if len(samples) == 0 {
		return index.TrainResult{}, fmt.Errorf("no training vectors")
	}
	if opts.NList <= 0 {
		return index.TrainResult{}, fmt.Errorf("nlist must be positive")
	}

	method := kmodes.InitRandom
	if opts.Init == "kmeans-plus-plus" {
		method = kmodes.InitKmeansPlusPlus
	}

	var st kmodes.State
	if !opts.TwoLevel || len(samples) < 30*opts.NList {
		st = kmodes.Cluster(samples, opts.NList, opts.MaxIter, method, nil)
	} else {
		st = kmodes.Cluster2Level(samples, opts.NList, opts.MaxIter, method, nil)
	}

	q := quantizer.New(st.Centroids)
	if err := q.Save(b.quantizerPath()); err != nil {
		return index.TrainResult{}, err
	}
	return index.TrainResult{
		Centroids: len(st.Centroids),
		Imbalance: float64(kmodes.ImbalanceFactor(st.Frequency)),
	}, nil
}

func (b *Backend) NewBuilder(_ context.Context) (index.Builder, error) {
	q, err := quantizer.Open(b.quantizerPath())
	if err != nil {
		return nil, fmt.Errorf("load quantizer: %w", err)
	}
	nlist := q.NList()
	il := invlists.NewArray(nlist)
	subIndex := 0
	for {
		if _, err := os.Stat(b.subIndexPath(subIndex)); os.IsNotExist(err) {
			break
		}
		subIndex++
	}
	return &builder{
		backend:  b,
		q:        q,
		nlist:    nlist,
		il:       il,
		idx:      ivf.New(q, il, 0),
		subIndex: subIndex,
	}, nil
}

func (b *Backend) NewSearcher(_ context.Context, threads int) (index.Searcher, error) {
	q, err := quantizer.Open(b.quantizerPath())
	if err != nil {
		return nil, fmt.Errorf("load quantizer: %w", err)
	}
	nlist := q.NList()

	var subs []invlists.InvertedLists
	var closers []io.Closer

	mainOD, err := invlists.LoadOnDisk(b.invlistsPath())
	if err != nil {
		return nil, fmt.Errorf("load invlists: %w", err)
	}
	if mainOD.NList() != nlist {
		mainOD.Close()
		return nil, fmt.Errorf("nlist mismatch: quantizer %d invlists %d", nlist, mainOD.NList())
	}
	subs = append(subs, mainOD)
	closers = append(closers, mainOD)

	for i := 0; ; i++ {
		subPath := b.subIndexPath(i)
		od, err := invlists.LoadOnDisk(subPath)
		if err != nil {
			if os.IsNotExist(err) {
				break
			}
			for _, c := range closers {
				c.Close()
			}
			return nil, fmt.Errorf("load shard %d: %w", i, err)
		}
		if od.NList() != nlist {
			od.Close()
			for _, c := range closers {
				c.Close()
			}
			return nil, fmt.Errorf("nlist mismatch in shard %d: %d vs %d", i, od.NList(), nlist)
		}
		subs = append(subs, od)
		closers = append(closers, od)
	}

	var il invlists.InvertedLists
	if len(subs) == 1 {
		il = subs[0]
	} else {
		v, err := invlists.NewVStack(subs...)
		if err != nil {
			for _, c := range closers {
				c.Close()
			}
			return nil, fmt.Errorf("vstack: %w", err)
		}
		il = v
	}

	return &searcher{idx: ivf.New(q, il, threads), closers: closers}, nil
}

func (b *Backend) Close() error { return nil }

type builder struct {
	backend  *Backend
	q        quantizer.Quantizer
	nlist    int
	il       invlists.InvertedLists
	idx      *ivf.Index
	subIndex int
	pending  bool
}

func (bl *builder) Add(_ context.Context, codes [][]byte, ids []uint64) error {
	if len(codes) == 0 {
		return nil
	}
	bl.idx.Add(codes, ids)
	bl.pending = true
	return nil
}

func (bl *builder) Flush(_ context.Context) error {
	if !bl.pending {
		return nil
	}
	subPath := bl.backend.subIndexPath(bl.subIndex)
	if err := invlists.Save(bl.il, subPath, bl.backend.codeSize, 1); err != nil {
		return err
	}
	bl.subIndex++
	bl.il = invlists.NewArray(bl.nlist)
	bl.idx = ivf.New(bl.q, bl.il, 0)
	bl.pending = false
	return nil
}

func (bl *builder) Commit(_ context.Context) error {
	return bl.commitShards(bl.subIndex)
}

func (bl *builder) Close() error { return nil }

func (bl *builder) commitShards(nSub int) error {
	if nSub == 0 {
		return nil
	}
	finalPath := bl.backend.invlistsPath()

	_, statErr := os.Stat(finalPath)

	if nSub == 1 && os.IsNotExist(statErr) {
		return os.Rename(bl.backend.subIndexPath(0), finalPath)
	}

	if os.IsNotExist(statErr) {
		return bl.mergeIntoNew(nSub, finalPath)
	}

	return nil
}

func (bl *builder) mergeIntoNew(nSub int, finalPath string) error {
	var subs []invlists.InvertedLists
	var closers []*invlists.OnDisk

	for i := range nSub {
		od, err := invlists.LoadOnDisk(bl.backend.subIndexPath(i))
		if err != nil {
			for _, c := range closers {
				c.Close()
			}
			return fmt.Errorf("merge: load shard %d: %w", i, err)
		}
		subs = append(subs, od)
		closers = append(closers, od)
	}

	v, err := invlists.NewVStack(subs...)
	if err != nil {
		for _, c := range closers {
			c.Close()
		}
		return fmt.Errorf("merge: vstack: %w", err)
	}
	tmpMerged := finalPath + ".merged"
	if err := invlists.Save(v, tmpMerged, bl.backend.codeSize, 3); err != nil {
		for _, c := range closers {
			c.Close()
		}
		return fmt.Errorf("merge: save: %w", err)
	}
	for _, c := range closers {
		c.Close()
	}
	if err := os.Rename(tmpMerged, finalPath); err != nil {
		return fmt.Errorf("merge: rename: %w", err)
	}
	for i := range nSub {
		os.Remove(bl.backend.subIndexPath(i))
	}
	return nil
}

type searcher struct {
	idx     *ivf.Index
	closers []io.Closer
}

func (s *searcher) Search(_ context.Context, descriptors [][]byte, k, nprobe int) []index.Neighbor {
	nn := s.idx.Search(descriptors, k, nprobe)
	out := make([]index.Neighbor, len(nn))
	for i, n := range nn {
		out[i] = index.Neighbor{Distance: n.Distance, ID: n.ID}
	}
	return out
}

func (s *searcher) Close() error {
	var firstErr error
	for _, c := range s.closers {
		if err := c.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}
