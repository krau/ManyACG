package migrate

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type SourceType string
type ArtworkStatus string
type StorageType string

const (
	SourceTypePixiv    SourceType = "pixiv"
	SourceTypeTwitter  SourceType = "twitter"
	SourceTypeBilibili SourceType = "bilibili"
	SourceTypeDanbooru SourceType = "danbooru"
	SourceTypeKemono   SourceType = "kemono"
	SourceTypeYandere  SourceType = "yandere"
	SourceTypeNhentai  SourceType = "nhentai"
)

const (
	ArtworkStatusCached  ArtworkStatus = "cached"
	ArtworkStatusPosting ArtworkStatus = "posting"
	ArtworkStatusPosted  ArtworkStatus = "posted"
)

const (
	StorageTypeWebdav   StorageType = "webdav"
	StorageTypeLocal    StorageType = "local"
	StorageTypeAlist    StorageType = "alist"
	StorageTypeTelegram StorageType = "telegram"
)

// ----- Artwork -----
type Artwork struct {
	// keep ObjectID as 24-hex string
	ID          MongoUUID  `gorm:"primaryKey;type:uuid" json:"id"`
	Title       string     `gorm:"type:text;not null;index:idx_artwork_title,sort:asc" json:"title"`
	Description string     `gorm:"type:text" json:"description"`
	R18         bool       `gorm:"not null;default:false" json:"r18"`
	CreatedAt   time.Time  `gorm:"not null;autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time  `gorm:"not null;autoUpdateTime" json:"updated_at"`
	SourceType  SourceType `gorm:"type:text;not null" json:"source_type"`
	SourceURL   string     `gorm:"type:text;not null;uniqueIndex" json:"source_url"`
	LikeCount   uint       `gorm:"not null;default:0" json:"like_count"`

	ArtistID MongoUUID `gorm:"type:uuid;index" json:"artist_id"`
	Artist   *Artist   `gorm:"foreignKey:ArtistID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL" json:"artist"`

	// many2many relationship with tags
	Tags []*Tag `gorm:"many2many:artwork_tags;constraint:OnDelete:CASCADE" json:"tags"`

	// one-to-many pictures
	Pictures []*Picture `gorm:"foreignKey:ArtworkID;constraint:OnDelete:CASCADE" json:"pictures"`
}

func (a *Artwork) BeforeCreate(tx *gorm.DB) (err error) {
	if a.ID == (MongoUUID{}) {
		a.ID = NewMongoUUID()
	}
	return nil
}

// ----- Artist -----
type Artist struct {
	ID       MongoUUID  `gorm:"primaryKey;type:uuid" json:"id"`
	Name     string     `gorm:"type:text;not null;index" json:"name"`
	Type     SourceType `gorm:"type:text;not null;index" json:"type"`
	UID      string     `gorm:"type:text;not null;index" json:"uid"`
	Username string     `gorm:"type:text;not null;index" json:"username"`

	// reverse relation
	Artworks []*Artwork `gorm:"foreignKey:ArtistID" json:"artworks"`
}

func (a *Artist) BeforeCreate(tx *gorm.DB) (err error) {
	if a.ID == (MongoUUID{}) {
		a.ID = NewMongoUUID()
	}
	return nil
}

// ----- Tag -----
type Tag struct {
	ID    MongoUUID   `gorm:"primaryKey;type:uuid" json:"id"`
	Name  string      `gorm:"type:text;not null;uniqueIndex" json:"name"`
	Alias []*TagAlias `gorm:"foreignKey:TagID;constraint:OnDelete:CASCADE" json:"alias"`

	// reverse relation via many2many
	Artworks []*Artwork `gorm:"many2many:artwork_tags" json:"artworks"`
}

type TagAlias struct {
	ID    MongoUUID `gorm:"primaryKey;type:uuid" json:"id"`
	TagID MongoUUID `gorm:"type:uuid;index" json:"tag_id"`
	Tag   *Tag      `gorm:"foreignKey:TagID;references:ID;constraint:OnDelete:CASCADE" json:"tag"`

	Alias string `gorm:"type:text;not null;index" json:"alias"`
}

func (t *Tag) BeforeCreate(tx *gorm.DB) (err error) {
	if t.ID == (MongoUUID{}) {
		t.ID = NewMongoUUID()
	}
	return nil
}

// ----- Picture -----
type Picture struct {
	ID        MongoUUID `gorm:"primaryKey;type:uuid" json:"id"`
	ArtworkID MongoUUID `gorm:"type:uuid;index" json:"artwork_id"`
	Artwork   *Artwork  `gorm:"foreignKey:ArtworkID;references:ID;constraint:OnDelete:CASCADE" json:"-"`

	Index     uint   `gorm:"not null;default:0;index:idx_picture_artwork_index,priority:1" json:"index"` // order within artwork
	Thumbnail string `gorm:"type:text" json:"thumbnail"`
	Original  string `gorm:"type:text;index" json:"original"`
	Width     uint   `json:"width"`
	Height    uint   `json:"height"`
	Phash     string `gorm:"type:varchar(18);index" json:"phash"` // phash
	ThumbHash string `gorm:"type:varchar(28)" json:"thumb_hash"`  // thumbhash

	TelegramInfo datatypes.JSONType[TelegramInfo] `gorm:"type:json" json:"telegram_info"` // original TelegramInfo struct as JSON
	StorageInfo  datatypes.JSONType[StorageInfo]  `gorm:"type:json" json:"storage_info"`  // StorageInfo as JSON

	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (p *Picture) BeforeCreate(tx *gorm.DB) (err error) {
	if p.ID == (MongoUUID{}) {
		p.ID = NewMongoUUID()
	}
	return nil
}

type DeletedRecord struct {
	ID        MongoUUID `gorm:"primaryKey;type:uuid" json:"id"`
	ArtworkID MongoUUID `gorm:"type:uuid;uniqueIndex" json:"artwork_id"`
	SourceURL string    `gorm:"type:text;not null;uniqueIndex" json:"source_url"`
	DeletedAt time.Time `gorm:"not null;autoCreateTime" json:"deleted_at"`
}

func (d *DeletedRecord) BeforeCreate(tx *gorm.DB) (err error) {
	if d.ID == (MongoUUID{}) {
		d.ID = NewMongoUUID()
	}
	return nil
}

type CachedArtworkData struct {
	ID          string     `json:"id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	R18         bool       `json:"r18"`
	SourceType  SourceType `json:"source_type"`
	SourceURL   string     `json:"source_url"`

	Artist   *CachedArtist    `json:"artist"`
	Tags     []string         `json:"tags"`
	Pictures []*CachedPicture `json:"pictures"`

	Version int `json:"version"` // for future schema changes
}

type CachedArtist struct {
	ID       string     `json:"id"`
	Name     string     `json:"name"`
	Type     SourceType `json:"type"`
	UID      string     `json:"uid"`
	Username string     `json:"username"`
}

type CachedPicture struct {
	ID        string `json:"id"`
	ArtworkID string `json:"artwork_id"`
	Index     uint   `json:"index"`
	Thumbnail string `json:"thumbnail"`
	Original  string `json:"original"`
	Hidden    bool   `json:"hidden"` // don't post when true

	Width     uint   `json:"width"`
	Height    uint   `json:"height"`
	Phash     string `json:"phash"`      // phash
	ThumbHash string `json:"thumb_hash"` // thumbhash
}

type TelegramInfo struct {
	PhotoFileID    string `json:"photo_file_id" bson:"photo_file_id"`
	DocumentFileID string `json:"document_file_id" bson:"document_file_id"`
	MessageID      int    `json:"message_id" bson:"message_id"`
	MediaGroupID   string `json:"media_group_id" bson:"media_group_id"`
}

type StorageInfo struct {
	Original *StorageDetail `json:"original" bson:"original"`
	Regular  *StorageDetail `json:"regular" bson:"regular"`
	Thumb    *StorageDetail `json:"thumb" bson:"thumb"`
}

type StorageDetail struct {
	Type StorageType `json:"type" bson:"type"`
	Path string      `json:"path" bson:"path"`
}

// ----- Cached Artworks -----
type CachedArtwork struct {
	ID        MongoUUID                              `gorm:"primaryKey;type:uuid" json:"id"`
	SourceURL string                                 `gorm:"type:text;uniqueIndex" json:"source_url"`
	CreatedAt time.Time                              `gorm:"autoCreateTime" json:"created_at"`
	Artwork   datatypes.JSONType[*CachedArtworkData] `gorm:"type:json" json:"artwork"`
	Status    ArtworkStatus                          `gorm:"type:text;index" json:"status"`
}

type ApiKey struct {
	ID          MongoUUID                   `gorm:"primaryKey;type:uuid" json:"id"`
	Key         string                      `gorm:"type:text;not null;uniqueIndex" json:"key"`
	Quota       int                         `gorm:"not null;default:0" json:"quota"`
	Used        int                         `gorm:"not null;default:0" json:"used"`
	Permissions datatypes.JSONSlice[string] `gorm:"type:json" json:"permissions"`
	Description string                      `gorm:"type:text" json:"description"`
}

func (a *ApiKey) BeforeCreate(tx *gorm.DB) (err error) {
	if a.ID == (MongoUUID{}) {
		a.ID = NewMongoUUID()
	}
	return nil
}

type User struct {
	ID         MongoUUID      `gorm:"primaryKey;type:uuid" json:"id"`
	Username   string         `gorm:"type:text;uniqueIndex" json:"username"`
	Password   string         `gorm:"type:text;not null" json:"password"`
	Email      *string        `gorm:"type:text;uniqueIndex" json:"email"`
	TelegramID *int64         `gorm:"type:bigint;uniqueIndex" json:"telegram_id"`
	Blocked    bool           `gorm:"not null;default:false;index" json:"blocked"`
	UpdatedAt  time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"deleted_at"`

	Favorites []*Artwork `gorm:"many2many:user_favorites;constraint:OnDelete:CASCADE" json:"favorites"`

	Settings datatypes.JSONType[*UserSettings] `gorm:"type:json" json:"settings"`
}

func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
	if u.ID == (MongoUUID{}) {
		u.ID = NewMongoUUID()
	}
	return nil
}

type UserSettings struct {
	Language string `json:"language"`
	Theme    string `json:"theme"`
	R18      bool   `json:"r18"`
}
