package repo

import (
	"context"
	// "github.com/krau/ManyACG/internal/infra/database"
)

// type defaultRepoImpl struct {
// 	admin         Admin
// 	apikey        APIKey
// 	artist        Artist
// 	artwork       Artwork
// 	deleted       DeletedRecord
// 	cachedArtwork CachedArtwork
// }

type Repositories interface {
	Admin() Admin
	APIKey() APIKey
	Artist() Artist
	Artwork() Artwork
	Tag() Tag
	Picture() Picture
	DeletedRecord() DeletedRecord
	CachedArtwork() CachedArtwork
	Transaction(ctx context.Context, fn func(repos Repositories) error) error
}

// func (r *defaultRepoImpl) Admin() Admin {
// 	return r.admin
// }

// func (r *defaultRepoImpl) APIKey() APIKey {
// 	return r.apikey
// }

// func (r *defaultRepoImpl) Artist() Artist {
// 	return r.artist
// }

// func (r *defaultRepoImpl) Artwork() Artwork {
// 	return r.artwork
// }

// func (r *defaultRepoImpl) DeletedRecord() DeletedRecord {
// 	return r.deleted
// }

// func (r *defaultRepoImpl) CachedArtwork() CachedArtwork {
// 	return r.cachedArtwork
// }

// var (
// 	defaultRepo *defaultRepoImpl
// 	defaultOnce sync.Once
// )

// func Default() Repositories {
// 	defaultOnce.Do(func() {
// 		defaultRepo = &defaultRepoImpl{
// 			admin:         database.Default(),
// 			apikey:        database.Default(),
// 			artist:        database.Default(),
// 			artwork:       database.Default(),
// 			deleted:       database.Default(),
// 			cachedArtwork: database.Default(),
// 		}
// 	})
// 	return defaultRepo
// }
