package entity

import (
	"time"

	"github.com/krau/ManyACG/internal/shared"
	"github.com/unvgo/ouid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

var _ shared.VideoLike = (*Video)(nil)

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

	TelegramInfo    datatypes.JSONType[shared.TelegramInfo]  `json:"telegram_info"`
	OriginalStorage datatypes.JSONType[shared.StorageDetail] `json:"original_storage"`

	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (v *Video) BeforeCreate(tx *gorm.DB) (err error) {
	if v.ID.IsZero() {
		v.ID = ouid.New()
	}
	return nil
}
// GetDuration implements [shared.VideoLike].
func (v *Video) GetDuration() uint {
	return v.Duration
}

// GetHeight implements [shared.VideoLike].
func (v *Video) GetHeight() uint {
	return v.Height
}

// GetIndex implements [shared.VideoLike].
func (v *Video) GetIndex() uint {
	return v.OrderIndex
}

// GetMimeType implements [shared.VideoLike].
func (v *Video) GetMimeType() string {
	return v.MimeType
}

// GetPoster implements [shared.VideoLike].
func (v *Video) GetPoster() string {
	return v.Poster
}

// GetStorageInfo implements [shared.VideoLike].
func (v *Video) GetOriginalStorage() shared.StorageDetail {
	return v.OriginalStorage.Data()
}

// GetTelegramInfo implements [shared.VideoLike].
func (v *Video) GetTelegramInfo() shared.TelegramInfo {
	return v.TelegramInfo.Data()
}

// GetURL implements [shared.VideoLike].
func (v *Video) GetURL() string {
	return v.URL
}

// GetWidth implements [shared.VideoLike].
func (v *Video) GetWidth() uint {
	return v.Width
}
