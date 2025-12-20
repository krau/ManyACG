package entity

import (
	"time"

	"github.com/krau/ManyACG/internal/shared"
	"github.com/unvgo/ouid"
	"gorm.io/datatypes"
)

type Video struct {
	ID         ouid.OUID `gorm:"primaryKey;type:uuid" json:"id"`
	ArtworkID  ouid.OUID `gorm:"type:uuid;index" json:"artwork_id"`
	Artwork    *Artwork  `gorm:"foreignKey:ArtworkID;references:ID;constraint:OnDelete:CASCADE" json:"-"`
	OrderIndex uint      `gorm:"column:order_index;not null;default:0;index:idx_video_artwork_index,priority:1" json:"index"`
	Poster     string    `gorm:"type:text" json:"poster"` // poster image URL
	Duration   uint      `json:"duration"`                // duration in ms
	Width      uint      `json:"width"`
	Height     uint      `json:"height"`
	MimeType   string    `gorm:"type:text" json:"mime_type"`
	URL        string    `gorm:"type:text;index" json:"url"` // video URL

	TelegramInfo datatypes.JSONType[shared.TelegramInfo] `json:"telegram_info"`
	StorageInfo  datatypes.JSONType[shared.StorageInfo]  `json:"storage_info"`

	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}
