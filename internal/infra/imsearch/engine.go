package imsearch

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"math"
	"sync"
	"time"

	"github.com/corona10/goimagehash"
	"github.com/krau/ManyACG/internal/infra/imsearch/config"
	"github.com/krau/ManyACG/internal/infra/imsearch/imdb"
	"github.com/krau/ManyACG/internal/infra/imsearch/index"
	"github.com/krau/ManyACG/internal/infra/imsearch/orb"
	"github.com/krau/ManyACG/pkg/log"
	"github.com/unvgo/ouid"
	"github.com/zeebo/blake3"
	_ "golang.org/x/image/webp"
)

type Hit struct {
	PictureID string
	ArtworkID string
	Score     float32
}

type SearchOpts struct {
	Distance uint32
	Count    int
	K        int
	NProbe   int
}

type Status struct {
	Images     int64
	Vectors    int64
	Unindexed  int64
	Trained    bool
	SearcherOK bool
	DataDir    string
}

type TrainOpts struct {
	// NList is number of IVF centroids; 0 = auto from vector count.
	NList int
	// Samples is number of descriptors to sample; 0 = 30 * NList (capped).
	Samples int
	MaxIter int
	// Force re-trains even if a quantizer already exists.
	Force bool
}

type Engine interface {
	Enabled() bool
	IndexPicture(ctx context.Context, pictureID, artworkID ouid.OUID, imageBytes []byte) error
	DeletePicture(ctx context.Context, pictureID ouid.OUID) error
	Search(ctx context.Context, imageBytes []byte, opts SearchOpts) ([]Hit, error)
	// Train builds the coarse quantizer from stored descriptors.
	Train(ctx context.Context, opts TrainOpts) error
	// Build indexes unindexed descriptor blobs into on-disk IVF invlists.
	Build(ctx context.Context) error
	// Rebuild force-retrains (if needed) and rebuilds the full IVF index.
	Rebuild(ctx context.Context) error
	Status(ctx context.Context) (Status, error)
	Close() error
}

type Config struct {
	Enable           bool
	DataDir          string
	Distance         int
	Count            int
	K                int
	NProbe           int
	NFeatures        int
	MaxHeight        int
	MaxWidth         int
	AutoBuild        bool
	BuildDebounceSec int
	MinMatches       int
	MinScore         float32
}

func DefaultConfig() Config {
	return Config{
		Enable:           false,
		DataDir:          "./data/imsearch",
		Distance:         64,
		Count:            10,
		K:                3,
		NProbe:           16,
		NFeatures:        500,
		MaxHeight:        1080,
		MaxWidth:         768,
		AutoBuild:        true,
		BuildDebounceSec: 2,
		MinMatches:       8,
		MinScore:         25,
	}
}

type engine struct {
	cfg Config
	db  *imdb.IMDB
	ext *orb.Extractor

	searchMu sync.RWMutex
	searcher index.Searcher

	// build coordination
	buildMu   sync.Mutex
	building  bool
	pending   bool
	buildTmr  *time.Timer
	closeOnce sync.Once
}

func Init(ctx context.Context, cfg Config) (Engine, error) {
	if !cfg.Enable {
		return &noopEngine{}, nil
	}
	if cfg.DataDir == "" {
		cfg.DataDir = DefaultConfig().DataDir
	}
	if cfg.BuildDebounceSec <= 0 {
		cfg.BuildDebounceSec = 2
	}
	if cfg.NFeatures <= 0 {
		cfg.NFeatures = 500
	}
	if cfg.MaxHeight <= 0 {
		cfg.MaxHeight = 1080
	}
	if cfg.MaxWidth <= 0 {
		cfg.MaxWidth = 768
	}

	m, err := imdb.Open(ctx, imdb.Options{
		ConfDir:  cfg.DataDir,
		WAL:      true,
		Cache:    true,
		CodeSize: 32,
	})
	if err != nil {
		return nil, fmt.Errorf("imsearch open: %w", err)
	}

	ext := orb.New(orb.Options{
		NFeatures: cfg.NFeatures,
		MaxSize:   orb.MaxSize{Height: cfg.MaxHeight, Width: cfg.MaxWidth},
	})

	e := &engine{cfg: cfg, db: m, ext: ext}
	// Try open searcher; ok if not trained yet.
	if s, _, err := m.OpenIndex(ctx, 0); err == nil {
		e.searcher = s
	} else {
		log.Warn("imsearch: index not ready yet", "err", err)
	}
	return e, nil
}

func (e *engine) Enabled() bool { return true }

func (e *engine) Close() error {
	var err error
	e.closeOnce.Do(func() {
		e.buildMu.Lock()
		if e.buildTmr != nil {
			e.buildTmr.Stop()
			e.buildTmr = nil
		}
		e.buildMu.Unlock()
		e.searchMu.Lock()
		if e.searcher != nil {
			_ = e.searcher.Close()
			e.searcher = nil
		}
		e.searchMu.Unlock()
		if e.ext != nil {
			_ = e.ext.Close()
		}
		if e.db != nil {
			err = e.db.Close()
		}
	})
	return err
}

func (e *engine) IndexPicture(ctx context.Context, pictureID, artworkID ouid.OUID, imageBytes []byte) error {
	if pictureID.IsZero() {
		return errors.New("imsearch: empty picture id")
	}
	if len(imageBytes) == 0 {
		return errors.New("imsearch: empty image bytes")
	}
	path := pictureID.Hex()
	artworkHex := ""
	if !artworkID.IsZero() {
		artworkHex = artworkID.Hex()
	}
	sum := blake3.Sum256(imageBytes)

	// Exact content hash: skip if another picture already indexed the same bytes.
	if id, ok, err := e.db.CheckHash(ctx, sum[:]); err != nil {
		return fmt.Errorf("imsearch check hash: %w", err)
	} else if ok {
		if img, err := e.db.GetImage(ctx, id); err == nil && img.Path != path {
			log.Debug("imsearch skip exact blake3 dup", "path", path, "existing", img.Path)
			return nil
		}
	}

	phashStr, dhashStr, err := computePerceptualHashes(imageBytes)
	if err != nil {
		log.Debug("imsearch perceptual hash failed, indexing without near-dup skip", "err", err)
	} else if existing, ok, err := e.db.FindByExactPerceptualHash(ctx, phashStr, dhashStr, path); err != nil {
		return fmt.Errorf("imsearch near-dup check: %w", err)
	} else if ok {
		log.Debug("imsearch skip exact phash+dhash dup", "path", path, "existing", existing)
		return nil
	}

	descs, err := e.ext.DetectBytes(imageBytes)
	if err != nil {
		return fmt.Errorf("imsearch extract: %w", err)
	}
	if len(descs) < 4 {
		return fmt.Errorf("imsearch: too few keypoints (%d)", len(descs))
	}
	// replace existing entry for same picture id
	_ = e.db.DeleteImageByPath(ctx, path)

	if _, err := e.db.AddImage(ctx, sum[:], path, artworkHex, phashStr, dhashStr, descs); err != nil {
		return fmt.Errorf("imsearch add: %w", err)
	}
	if e.cfg.AutoBuild {
		e.scheduleBuild()
	}
	return nil
}

func computePerceptualHashes(imageBytes []byte) (phash, dhash string, err error) {
	img, _, err := image.Decode(bytes.NewReader(imageBytes))
	if err != nil {
		return "", "", err
	}
	p, err := goimagehash.PerceptionHash(img)
	if err != nil {
		return "", "", err
	}
	d, err := goimagehash.DifferenceHash(img)
	if err != nil {
		return "", "", err
	}
	return p.ToString(), d.ToString(), nil
}

func (e *engine) DeletePicture(ctx context.Context, pictureID ouid.OUID) error {
	if pictureID.IsZero() {
		return nil
	}
	if err := e.db.DeleteImageByPath(ctx, pictureID.Hex()); err != nil {
		return err
	}
	// Reload searcher so deleted vectors are gone from on-disk invlists view.
	e.reloadSearcher(ctx)
	return nil
}

func (e *engine) Search(ctx context.Context, imageBytes []byte, opts SearchOpts) ([]Hit, error) {
	descs, err := e.ext.DetectBytes(imageBytes)
	if err != nil {
		return nil, fmt.Errorf("imsearch extract: %w", err)
	}
	if len(descs) == 0 {
		return nil, nil
	}
	sopts := e.searchOpts(opts)

	e.searchMu.RLock()
	s := e.searcher
	if s == nil {
		e.searchMu.RUnlock()
		e.reloadSearcher(ctx)
		e.searchMu.RLock()
		s = e.searcher
		if s == nil {
			e.searchMu.RUnlock()
			return nil, nil
		}
	}
	results, err := e.db.Search(ctx, s, descs, sopts)
	e.searchMu.RUnlock()
	if err != nil {
		return nil, err
	}
	out := make([]Hit, 0, len(results))
	for _, r := range results {
		out = append(out, Hit{PictureID: r.Path, ArtworkID: r.ArtworkID, Score: r.Score})
	}
	return out, nil
}

func (e *engine) searchOpts(o SearchOpts) config.SearchOptions {
	d := config.DefaultSearchOptions()
	if e.cfg.Distance > 0 {
		d.Distance = uint32(e.cfg.Distance)
	}
	if e.cfg.Count > 0 {
		d.Count = e.cfg.Count
	}
	if e.cfg.K > 0 {
		d.K = e.cfg.K
	}
	if e.cfg.NProbe > 0 {
		d.NProbe = e.cfg.NProbe
	}
	if e.cfg.MinMatches > 0 {
		d.MinMatches = e.cfg.MinMatches
	}
	if e.cfg.MinScore > 0 {
		d.MinScore = e.cfg.MinScore
	}
	if o.Distance > 0 {
		d.Distance = o.Distance
	}
	if o.Count > 0 {
		d.Count = o.Count
	}
	if o.K > 0 {
		d.K = o.K
	}
	if o.NProbe > 0 {
		d.NProbe = o.NProbe
	}
	return d
}

func (e *engine) Train(ctx context.Context, opts TrainOpts) error {
	if e.db.Backend().Trained(ctx) && !opts.Force {
		return nil
	}
	_, total, err := e.db.Count(ctx)
	if err != nil {
		return err
	}
	if total == 0 {
		return errors.New("imsearch train: no vectors stored (run add first)")
	}
	nlist := opts.NList
	if nlist <= 0 {
		nlist = autoNList(total)
	}
	samples := opts.Samples
	if samples <= 0 {
		samples = 30 * nlist
	}
	if int64(samples) > total {
		samples = int(total)
	}
	maxIter := opts.MaxIter
	if maxIter <= 0 {
		maxIter = 20
	}
	vecs, err := e.db.ExportVectors(ctx, samples)
	if err != nil {
		return err
	}
	if len(vecs) == 0 {
		return errors.New("imsearch train: export returned no vectors")
	}
	t0 := time.Now()
	res, err := e.db.TrainIndex(ctx, vecs, index.TrainOptions{
		NList: nlist, MaxIter: maxIter, Init: "kmeans-plus-plus", TwoLevel: true,
	})
	if err != nil {
		return fmt.Errorf("imsearch train: %w", err)
	}
	log.Info("imsearch train done",
		"nlist", nlist, "samples", len(vecs), "centroids", res.Centroids,
		"imbalance", res.Imbalance, "dur", time.Since(t0))
	return nil
}

func (e *engine) Build(ctx context.Context) error {
	if !e.db.Backend().Trained(ctx) {
		return errors.New("imsearch build: not trained (run train first)")
	}
	n, err := e.db.CountUnindexed(ctx)
	if err != nil {
		return err
	}
	if n == 0 {
		log.Info("imsearch build: nothing unindexed")
		e.reloadSearcher(ctx)
		return nil
	}
	t0 := time.Now()
	if err := e.db.BuildIndex(ctx, imdb.BuildOptions{}); err != nil {
		return fmt.Errorf("imsearch build: %w", err)
	}
	log.Info("imsearch build done", "unindexed", n, "dur", time.Since(t0))
	e.reloadSearcher(ctx)
	return nil
}

func (e *engine) Rebuild(ctx context.Context) error {
	if err := e.Train(ctx, TrainOpts{Force: true}); err != nil {
		return err
	}
	if err := e.forceUnindexed(ctx); err != nil {
		return err
	}
	return e.Build(ctx)
}

func (e *engine) Status(ctx context.Context) (Status, error) {
	images, vectors, err := e.db.Count(ctx)
	if err != nil {
		return Status{}, err
	}
	unindexed, err := e.db.CountUnindexed(ctx)
	if err != nil {
		return Status{}, err
	}
	e.searchMu.RLock()
	ok := e.searcher != nil
	e.searchMu.RUnlock()
	return Status{
		Images:     images,
		Vectors:    vectors,
		Unindexed:  unindexed,
		Trained:    e.db.Backend().Trained(ctx),
		SearcherOK: ok,
		DataDir:    e.cfg.DataDir,
	}, nil
}

func (e *engine) forceUnindexed(ctx context.Context) error {
	_, err := e.db.DB().SQL().ExecContext(ctx, `UPDATE vector_stats SET indexed = 0`)
	return err
}

func autoNList(totalVecs int64) int {
	if totalVecs <= 0 {
		return 64
	}
	n := int(math.Sqrt(float64(totalVecs)))
	if n < 64 {
		n = 64
	}
	if n > 16384 {
		n = 16384
	}
	return n
}

func (e *engine) scheduleBuild() {
	e.buildMu.Lock()
	defer e.buildMu.Unlock()
	if e.building {
		e.pending = true
		return
	}
	if e.buildTmr != nil {
		e.buildTmr.Reset(time.Duration(e.cfg.BuildDebounceSec) * time.Second)
		return
	}
	e.buildTmr = time.AfterFunc(time.Duration(e.cfg.BuildDebounceSec)*time.Second, func() {
		e.buildMu.Lock()
		e.buildTmr = nil
		e.buildMu.Unlock()
		e.runBuild()
	})
}

func (e *engine) runBuild() {
	e.buildMu.Lock()
	if e.building {
		e.buildMu.Unlock()
		return
	}
	e.building = true
	e.pending = false
	e.buildMu.Unlock()

	go func() {
		defer func() {
			e.buildMu.Lock()
			e.building = false
			need := e.pending
			e.pending = false
			e.buildMu.Unlock()
			if need {
				e.runBuild()
			}
		}()
		ctx := context.Background()
		// train if needed
		if !e.db.Backend().Trained(ctx) {
			if err := e.Rebuild(ctx); err != nil {
				log.Error("imsearch rebuild failed", "err", err)
			}
			return
		}
		n, _ := e.db.CountUnindexed(ctx)
		if n == 0 {
			return
		}
		t0 := time.Now()
		if err := e.db.BuildIndex(ctx, imdb.BuildOptions{}); err != nil {
			log.Error("imsearch build failed", "err", err)
			return
		}
		log.Info("imsearch build done", "unindexed", n, "dur", time.Since(t0))
		e.reloadSearcher(ctx)

		// compact when AutoBuild shards pile up.
		if localBE, ok := e.db.Backend().(interface {
			CountShards() int
			Compact() error
		}); ok && localBE.CountShards() >= 8 {
			t1 := time.Now()
			if err := localBE.Compact(); err != nil {
				log.Warn("imsearch compact failed", "err", err)
			} else {
				log.Info("imsearch compacted shards", "dur", time.Since(t1))
				e.reloadSearcher(ctx)
			}
		}
	}()
}

func (e *engine) reloadSearcher(ctx context.Context) {
	if err := e.db.ReloadCache(ctx); err != nil {
		log.Warn("imsearch cache reload", "err", err)
	}
	s, _, err := e.db.OpenIndex(ctx, 0)
	if err != nil {
		log.Warn("imsearch open index", "err", err)
		return
	}
	e.searchMu.Lock()
	old := e.searcher
	e.searcher = s
	e.searchMu.Unlock()
	// Close only after writers released the lock so no Search still holds old.
	if old != nil {
		_ = old.Close()
	}
}

type noopEngine struct{}

func (n *noopEngine) Enabled() bool { return false }
func (n *noopEngine) IndexPicture(context.Context, ouid.OUID, ouid.OUID, []byte) error {
	return nil
}
func (n *noopEngine) DeletePicture(context.Context, ouid.OUID) error { return nil }
func (n *noopEngine) Search(context.Context, []byte, SearchOpts) ([]Hit, error) {
	return nil, nil
}
func (n *noopEngine) Train(context.Context, TrainOpts) error { return nil }
func (n *noopEngine) Build(context.Context) error            { return nil }
func (n *noopEngine) Rebuild(context.Context) error          { return nil }
func (n *noopEngine) Status(context.Context) (Status, error) {
	return Status{}, nil
}
func (n *noopEngine) Close() error { return nil }
