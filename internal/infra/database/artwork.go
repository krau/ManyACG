package database

import (
	"context"

	"github.com/krau/ManyACG/internal/domain/entity/artwork"
	"github.com/krau/ManyACG/internal/domain/repo"
	"github.com/krau/ManyACG/internal/infra/database/po"
	"gorm.io/gorm"
)

type TxRepo struct {
	db *gorm.DB
}

func NewTxRepo(db *gorm.DB) *TxRepo {
	return &TxRepo{db: db}
}

func (r *TxRepo) WithTransaction(ctx context.Context, fn func(repos *repo.TransactionRepos) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(&repo.TransactionRepos{
			ArtworkRepo: NewArtworkRepo(tx),
			ArtistRepo:  NewArtistRepo(tx),
			TagRepo:     NewTagRepo(tx),
		})
	})
}

type artworkRepo struct {
	db *gorm.DB
}

func NewArtworkRepo(db *gorm.DB) repo.ArtworkRepo {
	return &artworkRepo{db: db}
}

func (r *artworkRepo) Save(ctx context.Context, artwork *artwork.Artwork) error {
	return r.db.WithContext(ctx).Save(po.ArtworkFromDomain(artwork)).Error
}

func (r *artworkRepo) FindByURL(ctx context.Context, sourceURL string) (*artwork.Artwork, error) {
	artwork, err := gorm.G[po.Artwork](r.db).Preload("Artist", func(db gorm.PreloadBuilder) error {
		db.Select("id")
		return nil
	}).Preload("Tags", func(db gorm.PreloadBuilder) error {
		db.Select("id")
		return nil
	}).Preload("Pictures", nil).Where("source_url = ?", sourceURL).First(ctx)
	if err != nil {
		return nil, err
	}
	return artwork.ToDomain(), nil
}
