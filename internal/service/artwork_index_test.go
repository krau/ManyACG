package service

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/krau/ManyACG/internal/infra/search"
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

func (r *indexRepairArtworkRepo) GetArtworksByIDs(ctx context.Context, ids []ouid.OUID) ([]*entity.Artwork, error) {
	wanted := make(map[ouid.OUID]bool, len(ids))
	for _, id := range ids {
		wanted[id] = true
	}
	var artworks []*entity.Artwork
	for _, artwork := range r.artworks {
		if wanted[artwork.ID] {
			artworks = append(artworks, artwork)
		}
	}
	return artworks, nil
}

type indexRepairSearcher struct {
	search.Searcher
	docs       map[string]*dto.ArtworkSearchDocument
	failure    error
	failLookup bool
	failAfter  string
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

func (s *indexRepairSearcher) AddDocuments(ctx context.Context, docs []*dto.ArtworkSearchDocument) error {
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

	submitted, err := serv.FixArtworkIndex(context.Background())
	if err != nil || submitted != 1002 {
		t.Fatalf("repair = (%d, %v), want (1002, nil)", submitted, err)
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
	submitted, err = serv.FixArtworkIndex(context.Background())
	if err != nil || submitted != 0 {
		t.Fatalf("repeat repair = (%d, %v), want (0, nil)", submitted, err)
	}
}

func TestFixArtworkIndexReportsPartialSubmission(t *testing.T) {
	serv, searcher, artworks := indexRepairFixture(t, 1001)
	failure := errors.New("search engine unavailable")
	searcher.failure = failure
	searcher.failLookup = true
	searcher.failAfter = artworks[999].ID.Hex()
	submitted, err := serv.FixArtworkIndex(context.Background())
	if submitted != 1000 || !errors.Is(err, failure) {
		t.Fatalf("repair = (%d, %v), want (1000, %v)", submitted, err, failure)
	}
	if _, ok := searcher.docs[artworks[1000].ID.Hex()]; ok {
		t.Fatal("repair indexed an artwork after lookup failed")
	}
}

func TestFixArtworkIndexDoesNotCountRejectedBatch(t *testing.T) {
	serv, searcher, artworks := indexRepairFixture(t, 1)
	failure := errors.New("index submission rejected")
	searcher.failure = failure
	submitted, err := serv.FixArtworkIndex(context.Background())
	if submitted != 0 || !errors.Is(err, failure) {
		t.Fatalf("repair = (%d, %v), want (0, %v)", submitted, err, failure)
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
	if submitted, err := serv.FixArtworkIndex(ctx); submitted != 0 || !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelled repair = (%d, %v)", submitted, err)
	}
}
