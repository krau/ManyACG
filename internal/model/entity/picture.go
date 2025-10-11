package entity

import (
	"time"

	"github.com/krau/ManyACG/internal/shared"
	"github.com/krau/ManyACG/pkg/objectuuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type Picture struct {
	ID        objectuuid.ObjectUUID `gorm:"primaryKey;type:uuid" json:"id"`
	ArtworkID objectuuid.ObjectUUID `gorm:"type:uuid;index" json:"artwork_id"`
	Artwork   *Artwork              `gorm:"foreignKey:ArtworkID;references:ID;constraint:OnDelete:CASCADE" json:"-"`

	OrderIndex uint   `gorm:"column:order_index;not null;default:0;index:idx_picture_artwork_index,priority:1" json:"index"`
	Thumbnail  string `gorm:"type:text" json:"thumbnail"`
	Original   string `gorm:"type:text;index" json:"original"`
	Width      uint   `json:"width"`
	Height     uint   `json:"height"`
	Phash      string `gorm:"type:varchar(32);index" json:"phash"` // phash
	ThumbHash  string `gorm:"type:varchar(32)" json:"thumb_hash"`  // thumbhash

	TelegramInfo datatypes.JSONType[shared.TelegramInfo] `json:"telegram_info"`
	StorageInfo  datatypes.JSONType[shared.StorageInfo]  `json:"storage_info"`

	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

// IsHide implements PictureLike.
func (p *Picture) IsHide() bool {
	return false
}

// GetIndex implements PictureLike.
func (p *Picture) GetIndex() uint {
	return p.OrderIndex
}

// GetOriginal implements PictureLike.
func (p *Picture) GetOriginal() string {
	return p.Original
}

// GetSize implements PictureLike.
func (p *Picture) GetSize() (width uint, height uint) {
	return p.Width, p.Height
}

// GetStorageInfo implements PictureLike.
func (p *Picture) GetStorageInfo() shared.StorageInfo {
	return p.StorageInfo.Data()
}

// GetTelegramInfo implements PictureLike.
func (p *Picture) GetTelegramInfo() shared.TelegramInfo {
	return p.TelegramInfo.Data()
}

// GetThumbnail implements PictureLike.
func (p *Picture) GetThumbnail() string {
	return p.Thumbnail
}

func (p *Picture) BeforeCreate(tx *gorm.DB) (err error) {
	if p.ID == objectuuid.Nil {
		p.ID = objectuuid.New()
	}
	return
}

type UgoiraMeta struct {
	ID        objectuuid.ObjectUUID `gorm:"primaryKey;type:uuid" json:"id"`
	ArtworkID objectuuid.ObjectUUID `gorm:"type:uuid;index" json:"artwork_id"`
	Artwork   *Artwork              `gorm:"foreignKey:ArtworkID;references:ID;constraint:OnDelete:CASCADE" json:"-"`

	OrderIndex uint                                      `gorm:"column:order_index;not null;default:0;index:idx_ugoira_artwork_index,priority:1" json:"index"`
	Data       datatypes.JSONType[shared.UgoiraMetaData] `json:"data"`

	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`

	OriginalStorage datatypes.JSONType[shared.StorageDetail] `json:"original_storage"`
	TelegramInfo    datatypes.JSONType[shared.TelegramInfo]  `json:"telegram_info"`
}

func (u *UgoiraMeta) BeforeCreate(tx *gorm.DB) (err error) {
	if u.ID == objectuuid.Nil {
		u.ID = objectuuid.New()
	}
	return
}
