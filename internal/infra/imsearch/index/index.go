package index

import "context"

type Neighbor struct {
	Distance uint32
	ID       uint64
	ImageID  int64 // 0 = unknown; resolve via vector-id map
}

type TrainOptions struct {
	NList    int    // number of coarse centroids (inverted lists)
	MaxIter  int    // clustering iterations
	Init     string // "random" | "kmeans-plus-plus"
	TwoLevel bool
}

type TrainResult struct {
	Centroids int
	Imbalance float64
}

type Builder interface {
	Add(ctx context.Context, codes [][]byte, ids []uint64) error
	Flush(ctx context.Context) error
	Commit(ctx context.Context) error
	Close() error
}

type Searcher interface {
	Search(ctx context.Context, descriptors [][]byte, k, nprobe int) []Neighbor
	Close() error
}

type Backend interface {
	Train(ctx context.Context, samples [][]byte, opts TrainOptions) (TrainResult, error)
	Trained(ctx context.Context) bool
	NewBuilder(ctx context.Context) (Builder, error)
	NewSearcher(ctx context.Context, threads int) (Searcher, error)
	// DeleteVectors removes global vector ids in [lo, hi] from the served index.
	DeleteVectors(ctx context.Context, lo, hi uint64) error
	Close() error
}
