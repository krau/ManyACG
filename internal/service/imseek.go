package service

import (
	"context"
	"errors"
	"io"
	"sort"

	"github.com/krau/ManyACG/internal/infra/imseek"
	"github.com/krau/ManyACG/internal/model/entity"
	"github.com/krau/ManyACG/internal/model/query"
	"github.com/krau/ManyACG/internal/shared"
	"github.com/krau/ManyACG/internal/shared/errs"
	"github.com/krau/ManyACG/pkg/log"
	"github.com/samber/oops"
	"github.com/unvgo/ouid"
)

type PictureSearchHit struct {
	Picture *entity.Picture
	Score   float32 // feature score (0 if phash-only)
	// Source: "phash", "feature", or "both"
	Source string
}

func (s *Service) SearchPicturesByImage(ctx context.Context, imageBytes []byte, phashDistance, limit int) ([]PictureSearchHit, error) {
	if limit <= 0 {
		limit = 10
	}
	if phashDistance < 0 {
		phashDistance = 10
	}

	type acc struct {
		pic     *entity.Picture
		score   float32
		phash   bool
		feature bool
	}
	byID := map[string]*acc{}

	// feature search first (fast IVF); results guide whether phash scan is needed.
	if s.imseek != nil && s.imseek.Enabled() {
		hits, err := s.imseek.Search(ctx, imageBytes, imseek.SearchOpts{Count: limit})
		if err != nil {
			log.Warn("imseek feature search failed", "err", err)
		} else {
			for _, h := range hits {
				pic, err := s.resolveImseekHit(ctx, h)
				if err != nil || pic == nil {
					continue
				}
				resolvedID := pic.ID.Hex()
				if a, ok := byID[resolvedID]; ok {
					a.feature = true
					if h.Score > a.score {
						a.score = h.Score
					}
					continue
				}
				byID[resolvedID] = &acc{pic: pic, score: h.Score, feature: true}
			}
		}
	}

	// phash scan only if feature search didn't fill enough results.
	if len(byID) < limit {
		hash, err := getPhashFromBytes(imageBytes)
		if err == nil && hash != "" {
			pics, err := s.QueryPicturesByPhash(ctx, query.PicturesPhash{
				Input:    hash,
				Distance: phashDistance,
				Limit:    limit,
			})
			if err != nil {
				log.Warn("phash search failed", "err", err)
			} else {
				for _, p := range pics {
					id := p.ID.Hex()
					if a, ok := byID[id]; ok {
						a.phash = true
					} else {
						byID[id] = &acc{pic: p, phash: true}
					}
				}
			}
		}
	}

	out := make([]PictureSearchHit, 0, len(byID))
	for _, a := range byID {
		src := "phash"
		if a.phash && a.feature {
			src = "both"
		} else if a.feature {
			src = "feature"
		}
		out = append(out, PictureSearchHit{Picture: a.pic, Score: a.score, Source: src})
	}
	// both > phash > feature; then by score; then by picture id for stability
	rank := map[string]int{"both": 0, "phash": 1, "feature": 2}
	sort.SliceStable(out, func(i, j int) bool {
		ri, rj := rank[out[i].Source], rank[out[j].Source]
		if ri != rj {
			return ri < rj
		}
		if out[i].Score != out[j].Score {
			return out[i].Score > out[j].Score
		}
		idi, idj := "", ""
		if out[i].Picture != nil {
			idi = out[i].Picture.ID.Hex()
		}
		if out[j].Picture != nil {
			idj = out[j].Picture.ID.Hex()
		}
		return idi < idj
	})
	if len(out) > limit {
		out = out[:limit]
	}
	return out, nil
}

func (s *Service) IndexPictureForImseek(ctx context.Context, pic *entity.Picture) error {
	if s.imseek == nil || !s.imseek.Enabled() || pic == nil {
		return nil
	}
	data, err := s.loadPictureBytesForImseek(ctx, pic)
	if err != nil {
		return err
	}
	return s.imseek.IndexPicture(ctx, pic.ID, pic.ArtworkID, data)
}

func (s *Service) resolveImseekHit(ctx context.Context, h imseek.Hit) (*entity.Picture, error) {
	oid, err := ouid.FromObjectIDHex(h.PictureID)
	if err != nil {
		return nil, err
	}
	pic, err := s.repos.Picture().GetPictureByID(ctx, oid)
	if err == nil {
		if pic.Artwork == nil {
			aw, awErr := s.repos.Artwork().GetArtworkByID(ctx, pic.ArtworkID)
			if awErr != nil {
				return nil, nil
			}
			pic.Artwork = aw
		}
		return pic, nil
	}
	if !errors.Is(err, errs.ErrRecordNotFound) {
		return nil, err
	}
	// Picture gone: fall back via stored artwork id.
	if h.ArtworkID == "" {
		return nil, nil
	}
	awOID, err := ouid.FromObjectIDHex(h.ArtworkID)
	if err != nil {
		return nil, nil
	}
	aw, err := s.repos.Artwork().GetArtworkByID(ctx, awOID)
	if err != nil {
		return nil, nil
	}
	if len(aw.Pictures) == 0 {
		return nil, nil
	}
	// Prefer first remaining picture; attach artwork for display.
	fallback := aw.Pictures[0]
	fallback.Artwork = aw
	return fallback, nil
}

func (s *Service) DeletePictureFromImseek(ctx context.Context, id ouid.OUID) error {
	if s.imseek == nil || !s.imseek.Enabled() {
		return nil
	}
	return s.imseek.DeletePicture(ctx, id)
}

func (s *Service) RebuildImseek(ctx context.Context) error {
	if s.imseek == nil || !s.imseek.Enabled() {
		return oops.New("imseek is not enabled")
	}
	return s.imseek.Rebuild(ctx)
}

func (s *Service) loadPictureBytesForImseek(ctx context.Context, pic *entity.Picture) ([]byte, error) {
	info := pic.StorageInfo.Data()
	var detail *shared.StorageDetail
	if info.Regular != nil && !info.Regular.IsZero() {
		detail = info.Regular
	} else if info.Original != nil && !info.Original.IsZero() {
		detail = info.Original
	}
	if detail == nil {
		return nil, oops.New("picture has no regular/original storage")
	}
	f, err := s.StorageGetFile(ctx, *detail)
	if err != nil {
		return nil, oops.Wrapf(err, "get storage file")
	}
	defer f.Close()
	return io.ReadAll(f)
}
