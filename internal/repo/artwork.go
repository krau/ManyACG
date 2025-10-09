package repo

import (
	"context"
	"sync"

	"github.com/krau/ManyACG/internal/model/converter"
	"github.com/krau/ManyACG/internal/model/dto"
	"github.com/krau/ManyACG/internal/model/entity"
	"github.com/krau/ManyACG/internal/model/query"
	"github.com/krau/ManyACG/pkg/objectuuid"
)

type Artwork interface {
	GetArtworkByID(ctx context.Context, id objectuuid.ObjectUUID) (*entity.Artwork, error)
	GetArtworkByURL(ctx context.Context, url string) (*entity.Artwork, error)
	CreateArtwork(ctx context.Context, artwork *entity.Artwork) (*objectuuid.ObjectUUID, error)
	UpdateArtworkByMap(ctx context.Context, id objectuuid.ObjectUUID, patch map[string]any) error
	UpdateArtworkTags(ctx context.Context, id objectuuid.ObjectUUID, tags []*entity.Tag) error
	UpdateArtworkPictures(ctx context.Context, id objectuuid.ObjectUUID, pictures []*entity.Picture) error
	ReorderArtworkPicturesByID(ctx context.Context, id objectuuid.ObjectUUID) error
	DeleteArtworkByID(ctx context.Context, id objectuuid.ObjectUUID) error
	QueryArtworks(ctx context.Context, que query.ArtworksDB) ([]*entity.Artwork, error)
	GetArtworksByIDs(ctx context.Context, ids []objectuuid.ObjectUUID) ([]*entity.Artwork, error)
}

type ArtworkWithEvent struct {
	inner    Artwork
	eventBus EventBus[*dto.ArtworkEventItem]
}

var _ Artwork = (*ArtworkWithEvent)(nil)

// CreateArtwork implements Artwork.
func (a *ArtworkWithEvent) CreateArtwork(ctx context.Context, artwork *entity.Artwork) (*objectuuid.ObjectUUID, error) {
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
func (a *ArtworkWithEvent) DeleteArtworkByID(ctx context.Context, id objectuuid.ObjectUUID) error {
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
func (a *ArtworkWithEvent) UpdateArtworkByMap(ctx context.Context, id objectuuid.ObjectUUID, patch map[string]any) error {
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
func (a *ArtworkWithEvent) UpdateArtworkPictures(ctx context.Context, id objectuuid.ObjectUUID, pictures []*entity.Picture) error {
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
func (a *ArtworkWithEvent) UpdateArtworkTags(ctx context.Context, id objectuuid.ObjectUUID, tags []*entity.Tag) error {
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
func (a *ArtworkWithEvent) GetArtworkByID(ctx context.Context, id objectuuid.ObjectUUID) (*entity.Artwork, error) {
	return a.inner.GetArtworkByID(ctx, id)
}

// GetArtworkByURL implements Artwork.
func (a *ArtworkWithEvent) GetArtworkByURL(ctx context.Context, url string) (*entity.Artwork, error) {
	return a.inner.GetArtworkByURL(ctx, url)
}

// GetArtworksByIDs implements Artwork.
func (a *ArtworkWithEvent) GetArtworksByIDs(ctx context.Context, ids []objectuuid.ObjectUUID) ([]*entity.Artwork, error) {
	return a.inner.GetArtworksByIDs(ctx, ids)
}

// QueryArtworks implements Artwork.
func (a *ArtworkWithEvent) QueryArtworks(ctx context.Context, que query.ArtworksDB) ([]*entity.Artwork, error) {
	return a.inner.QueryArtworks(ctx, que)
}

// ReorderArtworkPicturesByID implements Artwork.
func (a *ArtworkWithEvent) ReorderArtworkPicturesByID(ctx context.Context, id objectuuid.ObjectUUID) error {
	return a.inner.ReorderArtworkPicturesByID(ctx, id)
}

func NewArtworkWithEvent(inner Artwork, eventBus EventBus[*dto.ArtworkEventItem]) *ArtworkWithEvent {
	return &ArtworkWithEvent{
		inner:    inner,
		eventBus: eventBus,
	}
}

type artworkEventItem struct {
	typ EventType
	ent *dto.ArtworkEventItem
}

// using in transaction, record events but not publish immediately
type ArtworkWithRecorder struct {
	inner    Artwork
	recorder func(typ EventType, item *dto.ArtworkEventItem)
}

var _ Artwork = (*ArtworkWithRecorder)(nil)

func (a *ArtworkWithRecorder) CreateArtwork(ctx context.Context, artwork *entity.Artwork) (*objectuuid.ObjectUUID, error) {
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

func (a *ArtworkWithRecorder) DeleteArtworkByID(ctx context.Context, id objectuuid.ObjectUUID) error {
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

func (a *ArtworkWithRecorder) UpdateArtworkByMap(ctx context.Context, id objectuuid.ObjectUUID, patch map[string]any) error {
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

func (a *ArtworkWithRecorder) UpdateArtworkPictures(ctx context.Context, id objectuuid.ObjectUUID, pictures []*entity.Picture) error {
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

func (a *ArtworkWithRecorder) UpdateArtworkTags(ctx context.Context, id objectuuid.ObjectUUID, tags []*entity.Tag) error {
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

func (a *ArtworkWithRecorder) GetArtworkByID(ctx context.Context, id objectuuid.ObjectUUID) (*entity.Artwork, error) {
	return a.inner.GetArtworkByID(ctx, id)
}

func (a *ArtworkWithRecorder) GetArtworkByURL(ctx context.Context, url string) (*entity.Artwork, error) {
	return a.inner.GetArtworkByURL(ctx, url)
}

func (a *ArtworkWithRecorder) GetArtworksByIDs(ctx context.Context, ids []objectuuid.ObjectUUID) ([]*entity.Artwork, error) {
	return a.inner.GetArtworksByIDs(ctx, ids)
}

func (a *ArtworkWithRecorder) QueryArtworks(ctx context.Context, que query.ArtworksDB) ([]*entity.Artwork, error) {
	return a.inner.QueryArtworks(ctx, que)
}

func (a *ArtworkWithRecorder) ReorderArtworkPicturesByID(ctx context.Context, id objectuuid.ObjectUUID) error {
	return a.inner.ReorderArtworkPicturesByID(ctx, id)
}

type WithArtworkEventImpl struct {
	Tx          Transactional
	AdminRepo   Admin
	ApiKeyRepo  APIKey
	ArtistRepo  Artist
	ArtworkRepo Artwork
	TagRepo     Tag
	PictureRepo Picture
	DeletedRepo DeletedRecord
	CachedRepo  CachedArtwork

	ArtworkBus EventBus[*dto.ArtworkEventItem]
}

// APIKey implements Repositories.
func (r *WithArtworkEventImpl) APIKey() APIKey {
	return r.ApiKeyRepo
}

// Admin implements Repositories.
func (r *WithArtworkEventImpl) Admin() Admin {
	return r.AdminRepo
}

// Artist implements Repositories.
func (r *WithArtworkEventImpl) Artist() Artist {
	return r.ArtistRepo
}

// Artwork implements Repositories.
func (r *WithArtworkEventImpl) Artwork() Artwork {
	return r.ArtworkRepo
}

// CachedArtwork implements Repositories.
func (r *WithArtworkEventImpl) CachedArtwork() CachedArtwork {
	return r.CachedRepo
}

// DeletedRecord implements Repositories.
func (r *WithArtworkEventImpl) DeletedRecord() DeletedRecord {
	return r.DeletedRepo
}

// Picture implements Repositories.
func (r *WithArtworkEventImpl) Picture() Picture {
	return r.PictureRepo
}

// Tag implements Repositories.
func (r *WithArtworkEventImpl) Tag() Tag {
	return r.TagRepo
}

func (r *WithArtworkEventImpl) Transaction(ctx context.Context, fn func(repos Repositories) error) error {
	var events []artworkEventItem
	var eventsMu sync.Mutex
	err := r.Tx.Transaction(ctx, func(txRepos Repositories) error {
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

		txWrapper := &WithArtworkEventImpl{
			Tx:          txRepos,
			AdminRepo:   txRepos.Admin(),
			ApiKeyRepo:  txRepos.APIKey(),
			ArtistRepo:  txRepos.Artist(),
			ArtworkRepo: &ArtworkWithRecorder{inner: txRepos.Artwork(), recorder: rec},
			TagRepo:     txRepos.Tag(),
			PictureRepo: txRepos.Picture(),
			DeletedRepo: txRepos.DeletedRecord(),
			CachedRepo:  txRepos.CachedArtwork(),

			ArtworkBus: r.ArtworkBus,
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
