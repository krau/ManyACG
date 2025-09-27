package persist

import (
	"time"

	"github.com/krau/ManyACG/internal/artwork/domain"
	"github.com/krau/ManyACG/internal/shared"
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

	TelegramInfo datatypes.JSONType[TelegramInfo] `gorm:"type:json" json:"telegram_info"` // original TelegramInfo struct as JSON
	StorageInfo  datatypes.JSONType[StorageInfo]  `gorm:"type:json" json:"storage_info"`  // StorageInfo as JSON

	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

type TelegramInfo struct {
	PhotoFileID    string `json:"photo_file_id"`
	DocumentFileID string `json:"document_file_id"`
	MessageID      int    `json:"message_id"`
	MediaGroupID   string `json:"media_group_id"`
}

type StorageInfo struct {
	Original *StorageDetail `json:"original"`
	Regular  *StorageDetail `json:"regular"`
	Thumb    *StorageDetail `json:"thumb"`
}

type StorageDetail struct {
	Type shared.StorageType `json:"type"`
	Path string             `json:"path"`
}

func (p *Picture) BeforeCreate(tx *gorm.DB) (err error) {
	if p.ID == objectuuid.Nil {
		p.ID = objectuuid.New()
	}
	return
}

func fromDomainPictures(pics []domain.Picture) []*Picture {
	if pics == nil {
		panic("nil pictures slice")
	}

	out := make([]*Picture, 0, len(pics))
	for _, p := range pics {
		out = append(out, &Picture{
			ID:        p.ID,
			Index:     p.Index,
			Thumbnail: p.Thumbnail,
			Original:  p.Original,
			Width:     p.Width,
			Height:    p.Height,
			Phash:     p.Phash,
			ThumbHash: p.ThumbHash,
			TelegramInfo: func() datatypes.JSONType[TelegramInfo] {
				if p.TelegramInfo == nil {
					return datatypes.JSONType[TelegramInfo]{}
				}
				return datatypes.NewJSONType(TelegramInfo(*p.TelegramInfo))
			}(),
			StorageInfo: func() datatypes.JSONType[StorageInfo] {
				if p.StorageInfo == nil {
					return datatypes.JSONType[StorageInfo]{}
				}
				return datatypes.NewJSONType(StorageInfo{
					Original: func() *StorageDetail {
						if p.StorageInfo.Original == nil {
							return nil
						}
						return (*StorageDetail)(p.StorageInfo.Original)
					}(),
					Regular: func() *StorageDetail {
						if p.StorageInfo.Regular == nil {
							return nil
						}
						return (*StorageDetail)(p.StorageInfo.Regular)
					}(),
					Thumb: func() *StorageDetail {
						if p.StorageInfo.Thumb == nil {
							return nil
						}
						return (*StorageDetail)(p.StorageInfo.Thumb)
					}(),
				})

			}(),
		})
	}
	return out
}
