package service

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/krau/ManyACG/internal/infra/search"
	"github.com/krau/ManyACG/internal/model/converter"
	"github.com/krau/ManyACG/internal/model/dto"
	"github.com/krau/ManyACG/internal/model/entity"
	"github.com/krau/ManyACG/internal/repo"
	"github.com/unvgo/ouid"
)

type indexRepairRepo struct {
	repo.Repositories
	artwork *indexRepairArtworkRepo
}

func (r *indexRepairRepo) Artwork() repo.Artwork { return r.artwork }

type indexRepairArtworkRepo struct {
	repo.Artwork
	artworks []*entity.Artwork
}

func (r *indexRepairArtworkRepo) GetArtworkIDs(ctx context.Context, after ouid.OUID, limit int) ([]ouid.OUID, error) {
	var ids []ouid.OUID
	for _, artwork := range r.artworks {
		if artwork.ID.Hex() > after.Hex() {
			ids = append(ids, artwork.ID)
			if len(ids) == limit {
				break
			}
		}
	}
	return ids, nil
}

func (r *indexRepairArtworkRepo) GetArtworkSearchDocuments(ctx context.Context, ids []ouid.OUID) ([]*dto.ArtworkSearchDocument, error) {
	wanted := make(map[ouid.OUID]bool, len(ids))
	for _, id := range ids {
		wanted[id] = true
	}
	var docs []*dto.ArtworkSearchDocument
	for _, artwork := range r.artworks {
		if wanted[artwork.ID] {
			docs = append(docs, converter.EntityArtworkToSearchDocument(artwork))
		}
	}
	return docs, nil
}

type indexRepairSearcher struct {
	search.Searcher
	docs       map[string]*dto.ArtworkSearchDocument
	failure    error
	failLookup bool
	failAfter  string
	entered    chan struct{}
	resume     chan struct{}
}

func (s *indexRepairSearcher) GetMissingArtworkIDs(ctx context.Context, ids []ouid.OUID) ([]ouid.OUID, error) {
	if s.failLookup && ids[0].Hex() > s.failAfter {
		return nil, s.failure
	}
	var missing []ouid.OUID
	for _, id := range ids {
		if _, ok := s.docs[id.Hex()]; !ok {
			missing = append(missing, id)
		}
	}
	return missing, nil
}

func (s *indexRepairSearcher) AddDocumentsAndWait(ctx context.Context, docs []*dto.ArtworkSearchDocument) error {
	if s.entered != nil {
		select {
		case s.entered <- struct{}{}:
		default:
		}
		select {
		case <-s.resume:
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	if s.failure != nil && !s.failLookup {
		return s.failure
	}
	for _, doc := range docs {
		s.docs[doc.ID] = doc
	}
	return nil
}

func indexRepairFixture(t *testing.T, count int) (*Service, *indexRepairSearcher, []*entity.Artwork) {
	t.Helper()
	artworks := make([]*entity.Artwork, count)
	for i := range artworks {
		id, err := ouid.FromObjectIDHex(fmt.Sprintf("%024x", i+1))
		if err != nil {
			t.Fatal(err)
		}
		artworks[i] = &entity.Artwork{ID: id, Title: fmt.Sprintf("artwork-%d", i), R18: i%2 == 0}
	}
	searcher := &indexRepairSearcher{docs: make(map[string]*dto.ArtworkSearchDocument)}
	serv := &Service{
		repos:    &indexRepairRepo{artwork: &indexRepairArtworkRepo{artworks: artworks}},
		searcher: searcher,
	}
	return serv, searcher, artworks
}

func TestFixArtworkIndexRepairsMissingAcrossBatches(t *testing.T) {
	serv, searcher, artworks := indexRepairFixture(t, 1003)
	preserved := &dto.ArtworkSearchDocument{ID: artworks[0].ID.Hex(), Title: "already indexed"}
	searcher.docs[preserved.ID] = preserved

	progress, err := serv.FixArtworkIndex(context.Background())
	if err != nil || progress.Repaired != 1002 || progress.Scanned != 1003 || progress.Running || progress.Phase != ArtworkIndexCompleted {
		t.Fatalf("repair = (%+v, %v), want 1003 scanned and 1002 repaired", progress, err)
	}
	if searcher.docs[preserved.ID] != preserved {
		t.Fatal("repair replaced an existing document")
	}
	for _, artwork := range artworks[1:] {
		doc := searcher.docs[artwork.ID.Hex()]
		if doc == nil || doc.Title != artwork.Title || doc.R18 != artwork.R18 {
			t.Fatalf("missing or incorrect document for %s: %+v", artwork.ID.Hex(), doc)
		}
	}
	progress, err = serv.FixArtworkIndex(context.Background())
	if err != nil || progress.Repaired != 0 || progress.Scanned != 1003 {
		t.Fatalf("repeat repair = (%+v, %v), want 1003 scanned and none repaired", progress, err)
	}
}

func TestFixArtworkIndexReportsPartialRepair(t *testing.T) {
	serv, searcher, artworks := indexRepairFixture(t, 1001)
	failure := errors.New("search engine unavailable")
	searcher.failure = failure
	searcher.failLookup = true
	searcher.failAfter = artworks[999].ID.Hex()
	progress, err := serv.FixArtworkIndex(context.Background())
	if progress.Repaired != 1000 || progress.Scanned != 1000 || progress.Running || progress.Phase != ArtworkIndexFailed || !errors.Is(err, failure) {
		t.Fatalf("repair = (%+v, %v), want 1000 repaired and lookup error", progress, err)
	}
	if _, ok := searcher.docs[artworks[1000].ID.Hex()]; ok {
		t.Fatal("repair indexed an artwork after lookup failed")
	}
}

func TestFixArtworkIndexDoesNotCountRejectedBatch(t *testing.T) {
	serv, searcher, artworks := indexRepairFixture(t, 1)
	failure := errors.New("index submission rejected")
	searcher.failure = failure
	progress, err := serv.FixArtworkIndex(context.Background())
	if progress.Repaired != 0 || progress.Scanned != 1 || progress.Running || !errors.Is(err, failure) {
		t.Fatalf("repair = (%+v, %v), want zero repaired and batch error", progress, err)
	}
	if _, ok := searcher.docs[artworks[0].ID.Hex()]; ok {
		t.Fatal("rejected document was indexed")
	}
}

func TestFixArtworkIndexUnavailableAndCancelled(t *testing.T) {
	serv := &Service{}
	if _, err := serv.FixArtworkIndex(context.Background()); !errors.Is(err, search.ErrNotEnabled) {
		t.Fatalf("repair without searcher: %v", err)
	}
	serv, _, _ = indexRepairFixture(t, 1)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if progress, err := serv.FixArtworkIndex(ctx); progress.Repaired != 0 || progress.Running || !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelled repair = (%+v, %v)", progress, err)
	}
}

func TestFixArtworkIndexRejectsConcurrentRepairAndWaitsForBatch(t *testing.T) {
	serv, searcher, _ := indexRepairFixture(t, 1001)
	searcher.entered = make(chan struct{}, 1)
	searcher.resume = make(chan struct{})
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	type result struct {
		progress ArtworkIndexProgress
		err      error
	}
	done := make(chan result, 1)
	go func() {
		progress, err := serv.FixArtworkIndex(ctx)
		done <- result{progress, err}
	}()
	select {
	case <-searcher.entered:
	case <-ctx.Done():
		t.Fatal("repair did not reach indexing")
	}
	current := serv.ArtworkIndexProgress()
	if !current.Running || current.Scanned != 1000 || current.Repaired != 0 || current.Phase != ArtworkIndexIndexing {
		t.Fatalf("pending batch progress: %+v", current)
	}
	duplicate, err := serv.FixArtworkIndex(ctx)
	if !errors.Is(err, ErrArtworkIndexRepairRunning) || duplicate != current {
		t.Fatalf("duplicate repair = (%+v, %v), want current progress and running error", duplicate, err)
	}
	close(searcher.resume)
	finished := <-done
	if finished.err != nil || finished.progress.Scanned != 1001 || finished.progress.Repaired != 1001 || finished.progress.Running {
		t.Fatalf("completed repair: %+v", finished)
	}
	if current := serv.ArtworkIndexProgress(); current != finished.progress {
		t.Fatalf("stored progress %+v differs from final progress %+v", current, finished.progress)
	}
	if next, err := serv.FixArtworkIndex(ctx); err != nil || next.Repaired != 0 {
		t.Fatalf("next repair = (%+v, %v)", next, err)
	}
}

func TestFixArtworkIndexCancelledBatchReleasesJob(t *testing.T) {
	serv, searcher, _ := indexRepairFixture(t, 1)
	searcher.entered = make(chan struct{}, 1)
	searcher.resume = make(chan struct{})
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { _, err := serv.FixArtworkIndex(ctx); done <- err }()
	select {
	case <-searcher.entered:
	case <-time.After(5 * time.Second):
		cancel()
		t.Fatal("repair did not reach indexing")
	}
	cancel()
	if err := <-done; !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelled batch: %v", err)
	}
	if current := serv.ArtworkIndexProgress(); current.Running || current.Repaired != 0 || current.Phase != ArtworkIndexFailed {
		t.Fatalf("cancelled batch progress: %+v", current)
	}
	close(searcher.resume)
	if next, err := serv.FixArtworkIndex(context.Background()); err != nil || next.Repaired != 1 {
		t.Fatalf("repair after cancellation = (%+v, %v)", next, err)
	}
}
