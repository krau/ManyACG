package entity

import (
	"time"

	"github.com/krau/ManyACG/internal/shared"
	"github.com/unvgo/ouid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type UgoiraMeta struct {
	ID        ouid.OUID `gorm:"primaryKey;type:uuid" json:"id"`
	ArtworkID ouid.OUID `gorm:"type:uuid;index" json:"artwork_id"`
	Artwork   *Artwork  `gorm:"foreignKey:ArtworkID;references:ID;constraint:OnDelete:CASCADE" json:"-"`

	OrderIndex uint                                      `gorm:"column:order_index;not null;default:0;index:idx_ugoira_artwork_index,priority:1" json:"index"`
	Data       datatypes.JSONType[shared.UgoiraMetaData] `json:"data"`

	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`

	OriginalStorage datatypes.JSONType[shared.StorageDetail] `json:"original_storage"`
	TelegramInfo    datatypes.JSONType[shared.TelegramInfo]  `json:"telegram_info"`
}

// GetIndex implements shared.UgoiraMetaLike.
func (u *UgoiraMeta) GetIndex() uint {
	return u.OrderIndex
}

// GetOriginalStorage implements shared.UgoiraMetaLike.
func (u *UgoiraMeta) GetOriginalStorage() shared.StorageDetail {
	return u.OriginalStorage.Data()
}

// GetTelegramInfo implements shared.UgoiraMetaLike.
func (u *UgoiraMeta) GetTelegramInfo() shared.TelegramInfo {
	return u.TelegramInfo.Data()
}

// GetUgoiraMetaData implements shared.UgoiraMetaLike.
func (u *UgoiraMeta) GetUgoiraMetaData() shared.UgoiraMetaData {
	return u.Data.Data()
}

func (u *UgoiraMeta) BeforeCreate(tx *gorm.DB) (err error) {
	if u.ID.IsZero() {
		u.ID = ouid.New()
	}
	return
}
