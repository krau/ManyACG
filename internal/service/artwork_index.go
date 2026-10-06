package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/krau/ManyACG/internal/infra/search"
	"github.com/unvgo/ouid"
)

var ErrArtworkIndexRepairRunning = errors.New("artwork index repair is already running")

type ArtworkIndexPhase int

const (
	ArtworkIndexChecking ArtworkIndexPhase = iota
	ArtworkIndexLoading
	ArtworkIndexIndexing
	ArtworkIndexCompleted
	ArtworkIndexFailed
)

type ArtworkIndexProgress struct {
	Running   bool
	Scanned   int64
	Repaired  int64
	StartedAt time.Time
	Phase     ArtworkIndexPhase
}

func (s *Service) ArtworkIndexProgress() ArtworkIndexProgress {
	s.indexRepairMu.Lock()
	defer s.indexRepairMu.Unlock()
	return s.indexRepairProgress
}

func (s *Service) updateArtworkIndexProgress(scanned, repaired int, phase ArtworkIndexPhase) {
	s.indexRepairMu.Lock()
	defer s.indexRepairMu.Unlock()
	s.indexRepairProgress.Scanned += int64(scanned)
	s.indexRepairProgress.Repaired += int64(repaired)
	s.indexRepairProgress.Phase = phase
}

func (s *Service) FixArtworkIndex(ctx context.Context) (progress ArtworkIndexProgress, err error) {
	if s.searcher == nil {
		return progress, search.ErrNotEnabled
	}
	s.indexRepairMu.Lock()
	if s.indexRepairProgress.Running {
		progress = s.indexRepairProgress
		s.indexRepairMu.Unlock()
		return progress, ErrArtworkIndexRepairRunning
	}
	s.indexRepairProgress = ArtworkIndexProgress{Running: true, StartedAt: time.Now(), Phase: ArtworkIndexChecking}
	s.indexRepairMu.Unlock()
	defer func() {
		s.indexRepairMu.Lock()
		defer s.indexRepairMu.Unlock()
		s.indexRepairProgress.Running = false
		if err != nil {
			s.indexRepairProgress.Phase = ArtworkIndexFailed
		} else {
			s.indexRepairProgress.Phase = ArtworkIndexCompleted
		}
		progress = s.indexRepairProgress
	}()

	const batchSize = 1000
	var after ouid.OUID
	for {
		if err := ctx.Err(); err != nil {
			return progress, err
		}
		s.updateArtworkIndexProgress(0, 0, ArtworkIndexChecking)
		ids, err := s.repos.Artwork().GetArtworkIDs(ctx, after, batchSize)
		if err != nil {
			return progress, fmt.Errorf("get artwork ids failed: %w", err)
		}
		if len(ids) == 0 {
			return progress, nil
		}
		missing, err := s.searcher.GetMissingArtworkIDs(ctx, ids)
		if err != nil {
			return progress, fmt.Errorf("get missing artwork ids failed: %w", err)
		}
		s.updateArtworkIndexProgress(len(ids), 0, ArtworkIndexChecking)
		if len(missing) > 0 {
			s.updateArtworkIndexProgress(0, 0, ArtworkIndexLoading)
			docs, err := s.repos.Artwork().GetArtworkSearchDocuments(ctx, missing)
			if err != nil {
				return progress, fmt.Errorf("get artwork search documents failed: %w", err)
			}
			if len(docs) > 0 {
				s.updateArtworkIndexProgress(0, 0, ArtworkIndexIndexing)
				if err := s.searcher.AddDocumentsAndWait(ctx, docs); err != nil {
					return progress, fmt.Errorf("index artwork documents failed: %w", err)
				}
				s.updateArtworkIndexProgress(0, len(docs), ArtworkIndexChecking)
			}
		}
		after = ids[len(ids)-1]
	}
}
