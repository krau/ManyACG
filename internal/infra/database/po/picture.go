package po

import (
	"time"

	"github.com/krau/ManyACG/internal/common"
	"github.com/krau/ManyACG/internal/domain/entity/artwork"
	"github.com/krau/ManyACG/pkg/objectuuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type Picture struct {
	ID        objectuuid.ObjectUUID `gorm:"primaryKey;type:uuid" json:"id"`
	ArtworkID objectuuid.ObjectUUID `gorm:"type:uuid;index" json:"artwork_id"`
	Artwork   *Artwork              `gorm:"foreignKey:ArtworkID;references:ID;constraint:OnDelete:CASCADE" json:"-"`

	Index     uint   `gorm:"not null;default:0;index:idx_picture_artwork_index,priority:1" json:"index"` // order within artwork
	Thumbnail string `gorm:"type:text" json:"thumbnail"`
	Original  string `gorm:"type:text;index" json:"original"`
	Width     uint   `json:"width"`
	Height    uint   `json:"height"`
	Phash     string `gorm:"type:varchar(18);index" json:"phash"` // phash
	ThumbHash string `gorm:"type:varchar(28)" json:"thumb_hash"`  // thumbhash

	TelegramInfo datatypes.JSONType[common.TelegramInfo] `gorm:"type:json" json:"telegram_info"` // original TelegramInfo struct as JSON
	StorageInfo  datatypes.JSONType[common.StorageInfo]  `gorm:"type:json" json:"storage_info"`  // StorageInfo as JSON

	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (p *Picture) BeforeCreate(tx *gorm.DB) (err error) {
	if p.ID == objectuuid.Nil {
		p.ID = objectuuid.New()
	}
	return
}

func PiucturesFromDomain(pics []artwork.Picture) []*Picture {
	if pics == nil {
		panic("nil pictures slice")
	}

	out := make([]*Picture, 0, len(pics))
	for _, p := range pics {
		out = append(out, &Picture{
			ID:           p.ID,
			Index:        p.Index,
			Thumbnail:    p.Thumbnail,
			Original:     p.Original,
			Width:        p.Width,
			Height:       p.Height,
			Phash:        p.Phash,
			ThumbHash:    p.ThumbHash,
			TelegramInfo: datatypes.NewJSONType(*p.TelegramInfo),
			StorageInfo:  datatypes.NewJSONType(*p.StorageInfo),
		})
	}
	return out
}
