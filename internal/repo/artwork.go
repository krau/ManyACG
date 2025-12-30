package repo

import (
	"context"
	"sync"

	"github.com/krau/ManyACG/internal/model/converter"
	"github.com/krau/ManyACG/internal/model/dto"
	"github.com/krau/ManyACG/internal/model/entity"
	"github.com/krau/ManyACG/internal/model/query"
	"github.com/krau/ManyACG/internal/shared"
	"github.com/unvgo/ouid"
)

type Artwork interface {
	GetArtworkByID(ctx context.Context, id ouid.OUID) (*entity.Artwork, error)
	GetArtworkByURL(ctx context.Context, url string) (*entity.Artwork, error)
	CreateArtwork(ctx context.Context, artwork *entity.Artwork) (*ouid.OUID, error)
	UpdateArtworkByMap(ctx context.Context, id ouid.OUID, patch map[string]any) error
	UpdateArtworkTags(ctx context.Context, id ouid.OUID, tags []*entity.Tag) error
	UpdateArtworkPictures(ctx context.Context, id ouid.OUID, pictures []*entity.Picture) error
	ReorderArtworkPicturesByID(ctx context.Context, id ouid.OUID) error
	DeleteArtworkByID(ctx context.Context, id ouid.OUID) error
	QueryArtworks(ctx context.Context, que query.ArtworksDB) ([]*entity.Artwork, error)
	GetArtworksByIDs(ctx context.Context, ids []ouid.OUID) ([]*entity.Artwork, error)
	CountArtworks(ctx context.Context, r18 shared.R18Type) (int64, error)
}

type ArtworkWithEvent struct {
	inner    Artwork
	eventBus EventBus[*dto.ArtworkEventItem]
}

// CountArtworks implements Artwork.
func (a *ArtworkWithEvent) CountArtworks(ctx context.Context, r18 shared.R18Type) (int64, error) {
	return a.inner.CountArtworks(ctx, r18)
}

var _ Artwork = (*ArtworkWithEvent)(nil)

// CreateArtwork implements Artwork.
func (a *ArtworkWithEvent) CreateArtwork(ctx context.Context, artwork *entity.Artwork) (*ouid.OUID, error) {
	id, err := a.inner.CreateArtwork(ctx, artwork)
	if err != nil {
		return nil, err
	}
	ent, err := a.GetArtworkByID(ctx, *id)
	if err != nil {
		return nil, err
	}
	a.eventBus.Publish(EventTypeArtworkCreate, converter.EntityArtworkToDtoEventItem(ent))
	return id, nil
}

// DeleteArtworkByID implements Artwork.
func (a *ArtworkWithEvent) DeleteArtworkByID(ctx context.Context, id ouid.OUID) error {
	ent, err := a.GetArtworkByID(ctx, id)
	if err != nil {
		return err
	}
	if err := a.inner.DeleteArtworkByID(ctx, id); err != nil {
		return err
	}
	a.eventBus.Publish(EventTypeArtworkDelete, converter.EntityArtworkToDtoEventItem(ent))
	return nil
}

// UpdateArtworkByMap implements Artwork.
func (a *ArtworkWithEvent) UpdateArtworkByMap(ctx context.Context, id ouid.OUID, patch map[string]any) error {
	err := a.inner.UpdateArtworkByMap(ctx, id, patch)
	if err != nil {
		return err
	}
	ent, err := a.GetArtworkByID(ctx, id)
	if err != nil {
		return err
	}
	a.eventBus.Publish(EventTypeArtworkUpdate, converter.EntityArtworkToDtoEventItem(ent))
	return nil
}

// UpdateArtworkPictures implements Artwork.
func (a *ArtworkWithEvent) UpdateArtworkPictures(ctx context.Context, id ouid.OUID, pictures []*entity.Picture) error {
	err := a.inner.UpdateArtworkPictures(ctx, id, pictures)
	if err != nil {
		return err
	}
	ent, err := a.GetArtworkByID(ctx, id)
	if err != nil {
		return err
	}
	a.eventBus.Publish(EventTypeArtworkUpdate, converter.EntityArtworkToDtoEventItem(ent))
	return nil
}

// UpdateArtworkTags implements Artwork.
func (a *ArtworkWithEvent) UpdateArtworkTags(ctx context.Context, id ouid.OUID, tags []*entity.Tag) error {
	err := a.inner.UpdateArtworkTags(ctx, id, tags)
	if err != nil {
		return err
	}
	ent, err := a.GetArtworkByID(ctx, id)
	if err != nil {
		return err
	}
	a.eventBus.Publish(EventTypeArtworkUpdate, converter.EntityArtworkToDtoEventItem(ent))
	return nil
}

// GetArtworkByID implements Artwork.
func (a *ArtworkWithEvent) GetArtworkByID(ctx context.Context, id ouid.OUID) (*entity.Artwork, error) {
	return a.inner.GetArtworkByID(ctx, id)
}

// GetArtworkByURL implements Artwork.
func (a *ArtworkWithEvent) GetArtworkByURL(ctx context.Context, url string) (*entity.Artwork, error) {
	return a.inner.GetArtworkByURL(ctx, url)
}

// GetArtworksByIDs implements Artwork.
func (a *ArtworkWithEvent) GetArtworksByIDs(ctx context.Context, ids []ouid.OUID) ([]*entity.Artwork, error) {
	return a.inner.GetArtworksByIDs(ctx, ids)
}

// QueryArtworks implements Artwork.
func (a *ArtworkWithEvent) QueryArtworks(ctx context.Context, que query.ArtworksDB) ([]*entity.Artwork, error) {
	return a.inner.QueryArtworks(ctx, que)
}

// ReorderArtworkPicturesByID implements Artwork.
func (a *ArtworkWithEvent) ReorderArtworkPicturesByID(ctx context.Context, id ouid.OUID) error {
	return a.inner.ReorderArtworkPicturesByID(ctx, id)
}

func NewArtworkWithEvent(inner Artwork, eventBus EventBus[*dto.ArtworkEventItem]) *ArtworkWithEvent {
	return &ArtworkWithEvent{
		inner:    inner,
		eventBus: eventBus,
	}
}

type artworkEventItem struct {
	ent *dto.ArtworkEventItem
	typ EventType
}

// using in transaction, record events but not publish immediately
type ArtworkWithRecorder struct {
	inner    Artwork
	recorder func(typ EventType, item *dto.ArtworkEventItem)
}

// CountArtworks implements Artwork.
func (a *ArtworkWithRecorder) CountArtworks(ctx context.Context, r18 shared.R18Type) (int64, error) {
	return a.inner.CountArtworks(ctx, r18)
}

var _ Artwork = (*ArtworkWithRecorder)(nil)

func (a *ArtworkWithRecorder) CreateArtwork(ctx context.Context, artwork *entity.Artwork) (*ouid.OUID, error) {
	id, err := a.inner.CreateArtwork(ctx, artwork)
	if err != nil {
		return nil, err
	}
	ent, err := a.inner.GetArtworkByID(ctx, *id)
	if err != nil {
		return nil, err
	}
	if a.recorder != nil {
		a.recorder(EventTypeArtworkCreate, converter.EntityArtworkToDtoEventItem(ent))
	}
	return id, nil
}

func (a *ArtworkWithRecorder) DeleteArtworkByID(ctx context.Context, id ouid.OUID) error {
	ent, err := a.inner.GetArtworkByID(ctx, id)
	if err != nil {
		return err
	}
	if err := a.inner.DeleteArtworkByID(ctx, id); err != nil {
		return err
	}
	if a.recorder != nil {
		a.recorder(EventTypeArtworkDelete, converter.EntityArtworkToDtoEventItem(ent))
	}
	return nil
}

func (a *ArtworkWithRecorder) UpdateArtworkByMap(ctx context.Context, id ouid.OUID, patch map[string]any) error {
	if err := a.inner.UpdateArtworkByMap(ctx, id, patch); err != nil {
		return err
	}
	ent, err := a.inner.GetArtworkByID(ctx, id)
	if err != nil {
		return err
	}
	if a.recorder != nil {
		a.recorder(EventTypeArtworkUpdate, converter.EntityArtworkToDtoEventItem(ent))
	}
	return nil
}

func (a *ArtworkWithRecorder) UpdateArtworkPictures(ctx context.Context, id ouid.OUID, pictures []*entity.Picture) error {
	if err := a.inner.UpdateArtworkPictures(ctx, id, pictures); err != nil {
		return err
	}
	ent, err := a.inner.GetArtworkByID(ctx, id)
	if err != nil {
		return err
	}
	if a.recorder != nil {
		a.recorder(EventTypeArtworkUpdate, converter.EntityArtworkToDtoEventItem(ent))
	}
	return nil
}

func (a *ArtworkWithRecorder) UpdateArtworkTags(ctx context.Context, id ouid.OUID, tags []*entity.Tag) error {
	if err := a.inner.UpdateArtworkTags(ctx, id, tags); err != nil {
		return err
	}
	ent, err := a.inner.GetArtworkByID(ctx, id)
	if err != nil {
		return err
	}
	if a.recorder != nil {
		a.recorder(EventTypeArtworkUpdate, converter.EntityArtworkToDtoEventItem(ent))
	}
	return nil
}

func (a *ArtworkWithRecorder) GetArtworkByID(ctx context.Context, id ouid.OUID) (*entity.Artwork, error) {
	return a.inner.GetArtworkByID(ctx, id)
}

func (a *ArtworkWithRecorder) GetArtworkByURL(ctx context.Context, url string) (*entity.Artwork, error) {
	return a.inner.GetArtworkByURL(ctx, url)
}

func (a *ArtworkWithRecorder) GetArtworksByIDs(ctx context.Context, ids []ouid.OUID) ([]*entity.Artwork, error) {
	return a.inner.GetArtworksByIDs(ctx, ids)
}

func (a *ArtworkWithRecorder) QueryArtworks(ctx context.Context, que query.ArtworksDB) ([]*entity.Artwork, error) {
	return a.inner.QueryArtworks(ctx, que)
}

func (a *ArtworkWithRecorder) ReorderArtworkPicturesByID(ctx context.Context, id ouid.OUID) error {
	return a.inner.ReorderArtworkPicturesByID(ctx, id)
}

type WithArtworkEventImpl struct {
	inner      Repositories
	ArtworkBus EventBus[*dto.ArtworkEventItem]
}

func (r *WithArtworkEventImpl) Transaction(ctx context.Context, fn func(repos Repositories) error) error {
	var events []artworkEventItem
	var eventsMu sync.Mutex
	err := r.inner.Transaction(ctx, func(txRepos Repositories) error {
		rec := func(typ EventType, item *dto.ArtworkEventItem) {
			eventsMu.Lock()
			if item != nil {
				copied := *item
				events = append(events, artworkEventItem{typ: typ, ent: &copied})
			} else {
				events = append(events, artworkEventItem{typ: typ, ent: nil})
			}
			eventsMu.Unlock()
		}

		txWrapper := &txReposWithRecorder{
			inner:    txRepos,
			recorder: rec,
		}

		return fn(txWrapper)
	})
	if err != nil {
		return err
	}

	// transaction committed, now publish events
	for _, it := range events {
		if r.ArtworkBus != nil {
			r.ArtworkBus.Publish(it.typ, it.ent)
		}
	}

	return nil
}

var _ Repositories = (*WithArtworkEventImpl)(nil)

// NewWithArtworkEventImpl wraps base repositories with artwork event support.
// It keeps all existing behaviours, but decorates Artwork operations so they
// publish/search-index events via the provided EventBus.
func NewWithArtworkEventImpl(inner Repositories, bus EventBus[*dto.ArtworkEventItem]) *WithArtworkEventImpl {
	return &WithArtworkEventImpl{
		inner:      inner,
		ArtworkBus: bus,
	}
}

// Admin implements Repositories.
func (r *WithArtworkEventImpl) Admin() Admin {
	return r.inner.Admin()
}

// APIKey implements Repositories.
func (r *WithArtworkEventImpl) APIKey() APIKey {
	return r.inner.APIKey()
}

// Artist implements Repositories.
func (r *WithArtworkEventImpl) Artist() Artist {
	return r.inner.Artist()
}

// Artwork implements Repositories.
// Outside of explicit transactions, operations publish events immediately.
func (r *WithArtworkEventImpl) Artwork() Artwork {
	return NewArtworkWithEvent(r.inner.Artwork(), r.ArtworkBus)
}

// Tag implements Repositories.
func (r *WithArtworkEventImpl) Tag() Tag {
	return r.inner.Tag()
}

// Picture implements Repositories.
func (r *WithArtworkEventImpl) Picture() Picture {
	return r.inner.Picture()
}

// Ugoira implements Repositories.
func (r *WithArtworkEventImpl) Ugoira() Ugoira {
	return r.inner.Ugoira()
}

// Video implements Repositories.
func (r *WithArtworkEventImpl) Video() Video {
	return r.inner.Video()
}

// DeletedRecord implements Repositories.
func (r *WithArtworkEventImpl) DeletedRecord() DeletedRecord {
	return r.inner.DeletedRecord()
}

// CachedArtwork implements Repositories.
func (r *WithArtworkEventImpl) CachedArtwork() CachedArtwork {
	return r.inner.CachedArtwork()
}

// txReposWithRecorder is used inside a transaction to decorate Artwork with
// ArtworkWithRecorder while delegating other repos directly to the inner
// implementation, so events are recorded and published only after commit.
type txReposWithRecorder struct {
	inner    Repositories
	recorder func(typ EventType, item *dto.ArtworkEventItem)
}

func (t *txReposWithRecorder) Admin() Admin { return t.inner.Admin() }

func (t *txReposWithRecorder) APIKey() APIKey { return t.inner.APIKey() }

func (t *txReposWithRecorder) Artist() Artist { return t.inner.Artist() }

func (t *txReposWithRecorder) Artwork() Artwork {
	return &ArtworkWithRecorder{inner: t.inner.Artwork(), recorder: t.recorder}
}

func (t *txReposWithRecorder) Tag() Tag { return t.inner.Tag() }

func (t *txReposWithRecorder) Picture() Picture { return t.inner.Picture() }

func (t *txReposWithRecorder) Ugoira() Ugoira { return t.inner.Ugoira() }

func (t *txReposWithRecorder) Video() Video { return t.inner.Video() }

func (t *txReposWithRecorder) DeletedRecord() DeletedRecord { return t.inner.DeletedRecord() }

func (t *txReposWithRecorder) CachedArtwork() CachedArtwork { return t.inner.CachedArtwork() }

func (t *txReposWithRecorder) Transaction(ctx context.Context, fn func(repos Repositories) error) error {
	// Delegate nested transactions to inner, but keep the same recorder so all
	// events are collected and published after the outermost commit.
	return t.inner.Transaction(ctx, func(nested Repositories) error {
		nestedWrapper := &txReposWithRecorder{
			inner:    nested,
			recorder: t.recorder,
		}
		return fn(nestedWrapper)
	})
}
