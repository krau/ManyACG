package imdb

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"sort"
	"sync"

	"github.com/krau/ManyACG/internal/infra/imsearch/config"
	"github.com/krau/ManyACG/internal/infra/imsearch/db"
	"github.com/krau/ManyACG/internal/infra/imsearch/index"
	"github.com/krau/ManyACG/internal/infra/imsearch/index/local"
	"github.com/krau/ManyACG/internal/infra/imsearch/scoring"
)

type IMDB struct {
	db       *db.DB
	backend  index.Backend
	confDir  string
	codeSize int

	cache  bool
	mu     sync.RWMutex
	bounds []db.VectorBound
}

type Options struct {
	ConfDir  string
	WAL      bool
	Cache    bool
	CodeSize int
}

func Open(ctx context.Context, opts Options) (*IMDB, error) {
	if opts.CodeSize == 0 {
		opts.CodeSize = 32
	}
	if err := os.MkdirAll(opts.ConfDir, 0o755); err != nil {
		return nil, err
	}
	database, err := db.Open(opts.ConfDir, opts.WAL)
	if err != nil {
		return nil, err
	}
	if detected, err := database.GuessCodeSize(ctx); err != nil {
		database.Close()
		return nil, err
	} else if detected > 0 {
		opts.CodeSize = detected
	}
	backend := local.New(opts.ConfDir, opts.CodeSize)
	m := &IMDB{
		db:       database,
		backend:  backend,
		confDir:  opts.ConfDir,
		codeSize: opts.CodeSize,
		cache:    opts.Cache,
	}
	if opts.Cache {
		if err := m.loadVectorBounds(ctx); err != nil {
			backend.Close()
			database.Close()
			return nil, err
		}
	}
	return m, nil
}

func (m *IMDB) Close() error {
	if m.backend != nil {
		m.backend.Close()
	}
	if m.db != nil {
		return m.db.Close()
	}
	return nil
}

func (m *IMDB) Backend() index.Backend { return m.backend }
func (m *IMDB) DB() *db.DB             { return m.db }

func (m *IMDB) loadVectorBounds(ctx context.Context) error {
	bounds, err := m.db.GetAllVectorBounds(ctx)
	if err != nil {
		return err
	}
	m.mu.Lock()
	m.bounds = bounds
	m.mu.Unlock()
	return nil
}

func (m *IMDB) ReloadCache(ctx context.Context) error {
	if !m.cache {
		return nil
	}
	return m.loadVectorBounds(ctx)
}

func (m *IMDB) Count(ctx context.Context) (int64, int64, error) {
	return m.db.GetCount(ctx)
}

func (m *IMDB) CountUnindexed(ctx context.Context) (int64, error) {
	return m.db.CountImageUnindexed(ctx)
}

func (m *IMDB) CheckHash(ctx context.Context, hash []byte) (int64, bool, error) {
	return m.db.CheckImageHash(ctx, hash)
}

func (m *IMDB) AddImage(ctx context.Context, hash []byte, path, artworkID, phash, dhash string, descriptors [][]byte) (int64, error) {
	tx, err := m.db.Begin(ctx)
	if err != nil {
		return 0, err
	}
	id, err := m.db.AddImage(ctx, tx, hash, path, artworkID, phash, dhash)
	if err != nil {
		tx.Rollback()
		return 0, err
	}
	blob := flatten(descriptors)
	if err := m.db.AddVector(ctx, tx, id, blob); err != nil {
		tx.Rollback()
		return 0, err
	}
	if err := m.db.AddVectorStats(ctx, tx, id, int64(len(descriptors))); err != nil {
		tx.Rollback()
		return 0, err
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return id, nil
}

func (m *IMDB) FindByExactPerceptualHash(ctx context.Context, phash, dhash, excludePath string) (string, bool, error) {
	return m.db.FindByExactPerceptualHash(ctx, phash, dhash, excludePath)
}

func (m *IMDB) FindImageIDByPath(ctx context.Context, path string) (int64, error) {
	var id int64
	err := m.db.SQL().QueryRowContext(ctx, `SELECT id FROM image WHERE path = ?`, path).Scan(&id)
	return id, err
}

func (m *IMDB) DeleteImage(ctx context.Context, id int64) error {
	rng, err := m.db.GetVectorIDRange(ctx, id)
	hasRange := false
	switch {
	case err == nil:
		hasRange = rng.Lo > 0 && rng.Hi >= rng.Lo
	case errors.Is(err, sql.ErrNoRows):
	default:
		return err
	}
	if hasRange && m.backend != nil {
		if err := m.backend.DeleteVectors(ctx, uint64(rng.Lo), uint64(rng.Hi)); err != nil {
			return fmt.Errorf("delete image %d: purge index: %w", id, err)
		}
	}
	if _, err := m.db.DeleteImage(ctx, id); err != nil {
		return err
	}
	if m.cache {
		if err := m.loadVectorBounds(ctx); err != nil {
			return fmt.Errorf("delete image %d: reload cache: %w", id, err)
		}
	}
	return nil
}

func (m *IMDB) DeleteImageByPath(ctx context.Context, path string) error {
	id, err := m.FindImageIDByPath(ctx, path)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil // already gone
		}
		return err
	}
	return m.DeleteImage(ctx, id)
}

func (m *IMDB) GetImage(ctx context.Context, id int64) (*db.ImageInfo, error) {
	return m.db.GetImage(ctx, id)
}

func flatten(descriptors [][]byte) []byte {
	if len(descriptors) == 0 {
		return nil
	}
	cs := len(descriptors[0])
	out := make([]byte, 0, len(descriptors)*cs)
	for _, d := range descriptors {
		out = append(out, d...)
	}
	return out
}

func (m *IMDB) findImageID(ctx context.Context, vectorID uint64) (int64, error) {
	if !m.cache {
		return m.db.GetImageIDByVectorID(ctx, int64(vectorID))
	}
	m.mu.RLock()
	bounds := m.bounds
	m.mu.RUnlock()
	vid := int64(vectorID)
	i := sort.Search(len(bounds), func(i int) bool { return bounds[i].Total >= vid })
	if i >= len(bounds) {
		return 0, fmt.Errorf("vector id %d out of range", vectorID)
	}
	b := bounds[i]
	if vid <= b.Total-b.Count {
		return 0, fmt.Errorf("vector id %d orphan", vectorID)
	}
	return b.ImageID, nil
}

type SearchResult struct {
	Score     float32
	Path      string // picture id hex
	ArtworkID string // artwork id hex (may be empty for legacy rows)
}

func (m *IMDB) Search(ctx context.Context, searcher index.Searcher, descriptors [][]byte, opts config.SearchOptions) ([]SearchResult, error) {
	if len(descriptors) == 0 {
		return nil, nil
	}
	if searcher == nil {
		return nil, errors.New("search: no searcher")
	}
	neighbors := searcher.Search(ctx, descriptors, opts.K, opts.NProbe)
	return m.processNeighbors(ctx, neighbors, opts)
}

func (m *IMDB) processNeighbors(ctx context.Context, neighbors []index.Neighbor, opts config.SearchOptions) ([]SearchResult, error) {
	maxBits := m.codeSize * 8
	groups := make(map[int64][]float32)
	for _, n := range neighbors {
		if n.Distance > opts.Distance {
			continue
		}
		imgID := n.ImageID
		if imgID == 0 {
			var err error
			imgID, err = m.findImageID(ctx, n.ID)
			if err != nil {
				continue
			}
		}
		sim := 1.0 - float32(n.Distance)/float32(maxBits)
		groups[imgID] = append(groups[imgID], sim)
	}

	minMatches := opts.MinMatches
	if minMatches <= 0 {
		minMatches = 8
	}
	minScore := opts.MinScore
	if minScore <= 0 {
		minScore = 25
	}

	type scored struct {
		id    int64
		score float32
	}
	results := make([]scored, 0, len(groups))
	for id, scores := range groups {
		if len(scores) < minMatches {
			continue
		}
		s := 100 * scoring.Wilson(scores)
		if s < minScore {
			continue
		}
		results = append(results, scored{id: id, score: s})
	}
	sort.Slice(results, func(i, j int) bool { return results[i].score > results[j].score })
	if len(results) > opts.Count {
		results = results[:opts.Count]
	}

	out := make([]SearchResult, 0, len(results))
	for _, r := range results {
		img, err := m.db.GetImage(ctx, r.id)
		if err != nil {
			continue
		}
		out = append(out, SearchResult{Score: r.score, Path: img.Path, ArtworkID: img.ArtworkID})
	}
	return out, nil
}
