package po

import (
	"time"

	"github.com/krau/ManyACG/internal/common"
	"github.com/krau/ManyACG/internal/domain/entity/artwork"
	"github.com/krau/ManyACG/pkg/objectuuid"
)

type Artwork struct {
	// keep ObjectID as 24-hex string
	ID          objectuuid.ObjectUUID `gorm:"primaryKey;type:uuid" json:"id"`
	Title       string                `gorm:"type:text;not null;index:idx_artwork_title,sort:asc" json:"title"`
	Description string                `gorm:"type:text" json:"description"`
	R18         bool                  `gorm:"not null;default:false" json:"r18"`
	CreatedAt   time.Time             `gorm:"not null;autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time             `gorm:"not null;autoUpdateTime" json:"updated_at"`
	SourceType  common.SourceType     `gorm:"type:text;not null" json:"source_type"`
	SourceURL   string                `gorm:"type:text;not null;uniqueIndex" json:"source_url"`
	LikeCount   uint                  `gorm:"not null;default:0" json:"like_count"`

	ArtistID objectuuid.ObjectUUID `gorm:"type:uuid;index" json:"artist_id"`
	Artist   *Artist               `gorm:"foreignKey:ArtistID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL" json:"artist"`

	// many2many relationship with tags
	Tags []*Tag `gorm:"many2many:artwork_tags;constraint:OnDelete:CASCADE" json:"tags"`

	// one-to-many pictures
	Pictures []*Picture `gorm:"foreignKey:ArtworkID;constraint:OnDelete:CASCADE" json:"pictures"`
}

func ArtworkFromDomain(a *artwork.Artwork) *Artwork {
	if a == nil {
		panic("why you passing nil artwork")
	}
	return &Artwork{
		ID:          a.ID,
		Title:       a.Title,
		Description: a.Description,
		R18:         a.R18,
		SourceType:  a.SourceType,
		SourceURL:   a.SourceURL,
		LikeCount:   a.LikeCount,
		ArtistID:    a.ArtistID,
		Pictures:    PiucturesFromDomain(a.Pictures),
		Tags: func() []*Tag {
			if a.TagIDs == nil {
				return nil
			}
			tags := make([]*Tag, 0, a.TagIDs.Len())
			inputs := a.TagIDs.UnsafeValue()
			for _, id := range inputs {
				tags = append(tags, &Tag{ID: id})
			}
			return tags
		}(),
	}
}

func (a *Artwork) ToDomain() *artwork.Artwork {
	return &artwork.Artwork{
		ID:          a.ID,
		Title:       a.Title,
		Description: a.Description,
		R18:         a.R18,
		SourceType:  a.SourceType,
		SourceURL:   a.SourceURL,
		LikeCount:   a.LikeCount,
		ArtistID:    a.ArtistID,
		Pictures: func() []artwork.Picture {
			if a.Pictures == nil {
				return nil
			}
			pics := make([]artwork.Picture, 0, len(a.Pictures))
			for _, p := range a.Pictures {
				pics = append(pics, artwork.Picture{
					ID:        p.ID,
					ArtworkID: p.ArtworkID,
					Index:     p.Index,
					Thumbnail: p.Thumbnail,
					Original:  p.Original,
					Width:     p.Width,
					Height:    p.Height,
					Phash:     p.Phash,
					ThumbHash: p.ThumbHash,
					TelegramInfo: &common.TelegramInfo{
						PhotoFileID:    p.TelegramInfo.Data().PhotoFileID,
						DocumentFileID: p.TelegramInfo.Data().DocumentFileID,
						MessageID:      p.TelegramInfo.Data().MessageID,
						MediaGroupID:   p.TelegramInfo.Data().MediaGroupID,
					},
					StorageInfo: &common.StorageInfo{
						Original: p.StorageInfo.Data().Original,
						Regular:  p.StorageInfo.Data().Regular,
						Thumb:    p.StorageInfo.Data().Thumb,
					},
				})
			}
			return pics
		}(),
		TagIDs: func() *objectuuid.ObjectUUIDs {
			if a.Tags == nil {
				return nil
			}
			ids := make([]objectuuid.ObjectUUID, 0, len(a.Tags))
			for _, t := range a.Tags {
				ids = append(ids, t.ID)
			}
			return objectuuid.NewObjectUUIDs(ids...)
		}(),
	}
}
