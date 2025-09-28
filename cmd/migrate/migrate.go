package migrate

import (
	"context"
	"errors"
	"fmt"

	"github.com/krau/ManyACG/dao/collections"
	"github.com/krau/ManyACG/internal/infra/config"
	mongotypes "github.com/krau/ManyACG/types"
	_ "github.com/ncruces/go-sqlite3/embed"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	mongoopts "go.mongodb.org/mongo-driver/mongo/options"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type Option struct {
	MongoClient *mongo.Client
	GormDB      *gorm.DB
	Cfg         config.Config
}

func Run(ctx context.Context, opt *Option) error {
	if opt == nil {
		return errors.New("migrate option is nil")
	}
	db := opt.GormDB
	err := db.AutoMigrate(
		&Artist{},
		&Tag{},
		&TagAlias{},
		&Artwork{},
		&Picture{},
		&DeletedRecord{},
		&CachedArtwork{},
		&ApiKey{},
		&User{},
	)
	if err != nil {
		return err
	}
	mongoDB := opt.MongoClient.Database(opt.Cfg.Database.Database)
	collectionArtwork := mongoDB.Collection(collections.Artworks)
	collectionArtist := mongoDB.Collection(collections.Artists)
	collectionTag := mongoDB.Collection(collections.Tags)
	collectionPicture := mongoDB.Collection(collections.Pictures)
	collectionDeleted := mongoDB.Collection(collections.Deleted)
	collectionCachedArtwork := mongoDB.Collection(collections.CachedArtworks)
	collectionApiKey := mongoDB.Collection(collections.ApiKeys)
	collectionUser := mongoDB.Collection(collections.Users)
	collectionFavorite := mongoDB.Collection(collections.Favorites)
	if err := migrateArtists(ctx, collectionArtist, db); err != nil {
		return fmt.Errorf("migrate artists failed: %w", err)
	}
	if err := migrateTags(ctx, collectionTag, db); err != nil {
		return fmt.Errorf("migrate tags failed: %w", err)
	}
	if err := migrateArtworks(ctx, collectionArtwork, db); err != nil {
		return fmt.Errorf("migrate artworks failed: %w", err)
	}
	if err := migratePictures(ctx, collectionPicture, db); err != nil {
		return fmt.Errorf("migrate pictures failed: %w", err)
	}
	if err := migrateDeletedRecords(ctx, collectionDeleted, db); err != nil {
		return fmt.Errorf("migrate deleted records failed: %w", err)
	}
	if err := migrateCachedArtworks(ctx, collectionCachedArtwork, db); err != nil {
		return fmt.Errorf("migrate cached artworks failed: %w", err)
	}
	if err := migrateApiKeys(ctx, collectionApiKey, db); err != nil {
		return fmt.Errorf("migrate api keys failed: %w", err)
	}
	if err := migrateUsers(ctx, collectionUser, db); err != nil {
		return fmt.Errorf("migrate users failed: %w", err)
	}
	if err := migrateFavorites(ctx, collectionFavorite, db); err != nil {
		return fmt.Errorf("migrate favorites failed: %w", err)
	}

	return nil
}

func migrateArtists(ctx context.Context, collection *mongo.Collection, db *gorm.DB) error {
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)
	batch := make([]*Artist, 0, 500)
	for cursor.Next(ctx) {
		var artistModel mongotypes.ArtistModel
		if err := cursor.Decode(&artistModel); err != nil {
			return err
		}
		artist := &Artist{
			ID:       FromObjectID(artistModel.ID),
			Type:     SourceType(artistModel.Type),
			Username: artistModel.Username,
			Name:     artistModel.Name,
			UID:      artistModel.UID,
		}
		batch = append(batch, artist)
		if len(batch) >= 500 {
			db.Transaction(func(tx *gorm.DB) error {
				tx.CreateInBatches(batch, 250)
				tx.Commit()
				batch = batch[:0]
				return tx.Error
			})
		}
	}
	if err := cursor.Err(); err != nil {
		return err
	}
	if len(batch) > 0 {
		return db.CreateInBatches(batch, 250).Error
	}
	return nil
}

func migrateTags(ctx context.Context, collection *mongo.Collection, db *gorm.DB) error {
	seen := make(map[string]struct{})
	uniqueTags := make([]*Tag, 0)
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		var tagModel mongotypes.TagModel
		if err := cursor.Decode(&tagModel); err != nil {
			return err
		}
		if _, ok := seen[tagModel.Name]; !ok {
			seen[tagModel.Name] = struct{}{}
			aliases := make([]*TagAlias, 0, len(tagModel.Alias))
			aliasSeen := make(map[string]struct{})
			for _, aliasName := range tagModel.Alias {
				if _, ok := aliasSeen[aliasName]; !ok && aliasName != tagModel.Name {
					aliasSeen[aliasName] = struct{}{}
					aliases = append(aliases, &TagAlias{
						ID:    NewMongoUUID(),
						Alias: aliasName,
					})
				}
			}
			tag := &Tag{
				ID:    FromObjectID(tagModel.ID),
				Name:  tagModel.Name,
				Alias: aliases,
			}
			uniqueTags = append(uniqueTags, tag)
		}
	}
	return db.CreateInBatches(uniqueTags, 250).Error
}

func migrateArtworks(ctx context.Context, collection *mongo.Collection, db *gorm.DB) error {
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)
	batch := make([]*Artwork, 0, 500)
	for cursor.Next(ctx) {
		var artworkModel mongotypes.ArtworkModel
		if err := cursor.Decode(&artworkModel); err != nil {
			return err
		}
		tags := make([]*Tag, 0, len(artworkModel.Tags))
		for _, tagID := range artworkModel.Tags {
			tags = append(tags, &Tag{ID: FromObjectID(tagID)})
		}
		artwork := &Artwork{
			ID:          FromObjectID(artworkModel.ID),
			Title:       artworkModel.Title,
			Description: artworkModel.Description,
			R18:         artworkModel.R18,
			LikeCount:   artworkModel.LikeCount,
			CreatedAt:   artworkModel.CreatedAt.Time(),
			SourceType:  SourceType(artworkModel.SourceType),
			SourceURL:   artworkModel.SourceURL,
			ArtistID:    FromObjectID(artworkModel.ArtistID),
			Tags:        tags,
		}
		batch = append(batch, artwork)
		if len(batch) >= 500 {
			db.Transaction(func(tx *gorm.DB) error {
				tx.CreateInBatches(batch, 250)
				tx.Commit()
				batch = batch[:0]
				return tx.Error
			})
		}
	}
	if err := cursor.Err(); err != nil {
		return err
	}
	if len(batch) > 0 {
		return db.CreateInBatches(batch, 250).Error
	}
	return nil
}

func migratePictures(ctx context.Context, collection *mongo.Collection, db *gorm.DB) error {
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)
	batch := make([]*Picture, 0, 500)
	for cursor.Next(ctx) {
		var pictureModel mongotypes.PictureModel
		if err := cursor.Decode(&pictureModel); err != nil {
			return err
		}
		picture := &Picture{
			ID:        FromObjectID(pictureModel.ID),
			ArtworkID: FromObjectID(pictureModel.ArtworkID),
			Index:     pictureModel.Index,
			Thumbnail: pictureModel.Thumbnail,
			Original:  pictureModel.Original,
			Width:     pictureModel.Width,
			Height:    pictureModel.Height,
			Phash:     pictureModel.Hash,
			ThumbHash: pictureModel.ThumbHash,
			TelegramInfo: datatypes.NewJSONType(TelegramInfo{
				PhotoFileID:    pictureModel.TelegramInfo.PhotoFileID,
				DocumentFileID: pictureModel.TelegramInfo.DocumentFileID,
				MessageID:      pictureModel.TelegramInfo.MessageID,
				MediaGroupID:   pictureModel.TelegramInfo.MediaGroupID,
			}),
			StorageInfo: datatypes.NewJSONType(StorageInfo{
				Original: func() *StorageDetail {
					if pictureModel.StorageInfo.Original != nil {
						return &StorageDetail{
							StorageType(pictureModel.StorageInfo.Original.Type),
							pictureModel.StorageInfo.Original.Path,
						}
					}
					return nil
				}(),
				Regular: func() *StorageDetail {
					if pictureModel.StorageInfo.Regular != nil {
						return &StorageDetail{
							StorageType(pictureModel.StorageInfo.Regular.Type),
							pictureModel.StorageInfo.Regular.Path,
						}
					}
					return nil
				}(),
				Thumb: func() *StorageDetail {
					if pictureModel.StorageInfo.Thumb != nil {
						return &StorageDetail{
							StorageType(pictureModel.StorageInfo.Thumb.Type),
							pictureModel.StorageInfo.Thumb.Path,
						}
					}
					return nil
				}(),
			}),
		}
		batch = append(batch, picture)
		if len(batch) >= 500 {
			db.Transaction(func(tx *gorm.DB) error {
				tx.CreateInBatches(batch, 250)
				tx.Commit()
				batch = batch[:0]
				return tx.Error
			})
		}
	}
	if err := cursor.Err(); err != nil {
		return err
	}
	if len(batch) > 0 {
		return db.CreateInBatches(batch, 250).Error
	}
	return nil
}

func migrateDeletedRecords(ctx context.Context, collection *mongo.Collection, db *gorm.DB) error {
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)
	deletes := make([]*DeletedRecord, 0)
	for cursor.Next(ctx) {
		var deletedModel mongotypes.DeletedModel
		if err := cursor.Decode(&deletedModel); err != nil {
			return err
		}
		deleted := &DeletedRecord{
			ID:        FromObjectID(deletedModel.ID),
			ArtworkID: FromObjectID(deletedModel.ArtworkID),
			SourceURL: deletedModel.SourceURL,
			DeletedAt: deletedModel.DeletedAt.Time(),
		}
		deletes = append(deletes, deleted)
	}
	return db.CreateInBatches(deletes, 250).Error
}

func migrateCachedArtworks(ctx context.Context, collection *mongo.Collection, db *gorm.DB) error {
	findOpt := mongoopts.Find()
	findOpt.SetSort(bson.D{{Key: "source_url", Value: 1}, {Key: "created_at", Value: -1}})
	cursor, err := collection.Find(ctx, bson.M{}, findOpt)
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)
	batch := make([]*CachedArtwork, 0, 500)
	var lastSourceURL string
	for cursor.Next(ctx) {
		var cachedModel mongotypes.CachedArtworksModel
		if err := cursor.Decode(&cachedModel); err != nil {
			return err
		}
		if cachedModel.SourceURL == lastSourceURL {
			// skip older duplicate
			continue
		}
		lastSourceURL = cachedModel.SourceURL
		data := &CachedArtworkData{
			Version:     1,
			ID:          cachedModel.Artwork.ID,
			Title:       cachedModel.Artwork.Title,
			Description: cachedModel.Artwork.Description,
			R18:         cachedModel.Artwork.R18,
			SourceType:  SourceType(cachedModel.Artwork.SourceType),
			SourceURL:   cachedModel.Artwork.SourceURL,

			Artist: &CachedArtist{
				ID:       cachedModel.Artwork.Artist.ID,
				Name:     cachedModel.Artwork.Artist.Name,
				Type:     SourceType(cachedModel.Artwork.Artist.Type),
				UID:      cachedModel.Artwork.Artist.UID,
				Username: cachedModel.Artwork.Artist.Username,
			},
			Tags: cachedModel.Artwork.Tags,
			Pictures: func() []*CachedPicture {
				pics := make([]*CachedPicture, len(cachedModel.Artwork.Pictures))
				for i, pic := range cachedModel.Artwork.Pictures {
					pics[i] = &CachedPicture{
						ID:        pic.ID,
						ArtworkID: pic.ArtworkID,
						Index:     pic.Index,
						Thumbnail: pic.Thumbnail,
						Original:  pic.Original,
						Hidden:    false,
						Width:     pic.Width,
						Height:    pic.Height,
						Phash:     pic.Hash,
						ThumbHash: pic.ThumbHash,
					}
				}
				return pics
			}(),
		}
		cached := &CachedArtwork{
			ID:        FromObjectID(cachedModel.ID),
			SourceURL: cachedModel.SourceURL,
			CreatedAt: cachedModel.CreatedAt.Time(),
			Status:    ArtworkStatus(cachedModel.Status),
			Artwork:   datatypes.NewJSONType(sanitizeArtworkData(data)),
		}
		batch = append(batch, cached)
		if len(batch) >= 500 {
			db.Transaction(func(tx *gorm.DB) error {
				tx.CreateInBatches(batch, 250)
				tx.Commit()
				batch = batch[:0]
				return tx.Error
			})
		}
	}
	if err := cursor.Err(); err != nil {
		return err
	}
	if len(batch) > 0 {
		return db.CreateInBatches(batch, 250).Error
	}
	return nil
}

func migrateApiKeys(ctx context.Context, collection *mongo.Collection, db *gorm.DB) error {
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		var apiKeyModel mongotypes.ApiKeyModel
		if err := cursor.Decode(&apiKeyModel); err != nil {
			return err
		}
		apiKey := &ApiKey{
			ID:          FromObjectID(apiKeyModel.ID),
			Key:         apiKeyModel.Key,
			Quota:       apiKeyModel.Quota,
			Used:        apiKeyModel.Used,
			Permissions: datatypes.JSONSlice[string](make([]string, len(apiKeyModel.Permissions))),
			Description: apiKeyModel.Description,
		}
		for i, perm := range apiKeyModel.Permissions {
			apiKey.Permissions[i] = string(perm)
		}
		if err := db.Create(apiKey).Error; err != nil {
			return err
		}
	}
	return nil
}

func migrateUsers(ctx context.Context, collection *mongo.Collection, db *gorm.DB) error {
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		var userModel mongotypes.UserModel
		if err := cursor.Decode(&userModel); err != nil {
			return err
		}
		user := &User{
			ID:       FromObjectID(userModel.ID),
			Username: userModel.Username,
			Password: userModel.Password,
			Email: func() *string {
				if userModel.Email == "" {
					return nil
				}
				return &userModel.Email
			}(),
			TelegramID: func() *int64 {
				if userModel.TelegramID == 0 {
					return nil
				}
				return &userModel.TelegramID
			}(),
			Blocked:   userModel.Blocked,
			UpdatedAt: userModel.UpdatedAt.Time(),
			DeletedAt: func() gorm.DeletedAt {
				if userModel.DeletedAt.Time().IsZero() {
					return gorm.DeletedAt{}
				}
				return gorm.DeletedAt{
					Time:  userModel.DeletedAt.Time(),
					Valid: true,
				}
			}(),
			Settings: datatypes.NewJSONType(&UserSettings{
				Language: userModel.Settings.Language,
				Theme:    userModel.Settings.Theme,
				R18:      userModel.Settings.R18,
			}),
		}
		if err := db.Create(user).Error; err != nil {
			return err
		}
	}
	return nil
}

func migrateFavorites(ctx context.Context, collection *mongo.Collection, db *gorm.DB) error {
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		var favModel mongotypes.FavoriteModel
		if err := cursor.Decode(&favModel); err != nil {
			return err
		}
		if err := db.Transaction(func(tx *gorm.DB) error {
			var user User
			if err := tx.Where("id = ?", FromObjectID(favModel.UserID).Hex()).First(&user).Error; err != nil {
				return err
			}
			var artwork Artwork
			if err := tx.Where("id = ?", FromObjectID(favModel.ArtworkID).Hex()).First(&artwork).Error; err != nil {
				return err
			}
			if err := tx.Model(&user).Association("Favorites").Append(&artwork); err != nil {
				return err
			}
			return nil
		}); err != nil {
			return err
		}
	}
	return nil
}
