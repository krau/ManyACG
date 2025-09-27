package persist

import (
	"time"

	"github.com/krau/ManyACG/internal/artwork/domain"
	"github.com/krau/ManyACG/internal/shared"
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
	SourceType  shared.SourceType     `gorm:"type:text;not null" json:"source_type"`
	SourceURL   string                `gorm:"type:text;not null;uniqueIndex" json:"source_url"`
	LikeCount   uint                  `gorm:"not null;default:0" json:"like_count"`

	ArtistID objectuuid.ObjectUUID `gorm:"type:uuid;index" json:"artist_id"`
	Artist   *Artist               `gorm:"foreignKey:ArtistID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL" json:"artist"`

	// many2many relationship with tags
	Tags []*Tag `gorm:"many2many:artwork_tags;constraint:OnDelete:CASCADE" json:"tags"`

	// one-to-many pictures
	Pictures []*Picture `gorm:"foreignKey:ArtworkID;constraint:OnDelete:CASCADE" json:"pictures"`
}

func fromDomain(a *domain.Artwork) *Artwork {
	if a == nil {
		panic("why you passing nil artwork")
	}
	return &Artwork{
		ID:          a.ID.Value(),
		Title:       a.Title,
		Description: a.Description,
		R18:         a.R18,
		SourceType:  a.SourceType,
		SourceURL:   a.SourceURL,
		LikeCount:   a.LikeCount,
		ArtistID:    a.ArtistID.Value(),
		Pictures:    fromDomainPictures(a.Pictures),
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

func (a *Artwork) toDomain() *domain.Artwork {
	return &domain.Artwork{
		ID:          domain.NewArtworkID(a.ID),
		Title:       a.Title,
		Description: a.Description,
		R18:         a.R18,
		SourceType:  a.SourceType,
		SourceURL:   a.SourceURL,
		LikeCount:   a.LikeCount,
		ArtistID:    domain.NewArtistID(a.ArtistID),
		Pictures: func() []domain.Picture {
			if a.Pictures == nil {
				return nil
			}
			pics := make([]domain.Picture, 0, len(a.Pictures))
			for _, p := range a.Pictures {
				pics = append(pics, domain.Picture{
					ID:        p.ID,
					ArtworkID: domain.NewArtworkID(p.ArtworkID),
					Index:     p.Index,
					Thumbnail: p.Thumbnail,
					Original:  p.Original,
					Width:     p.Width,
					Height:    p.Height,
					Phash:     p.Phash,
					ThumbHash: p.ThumbHash,
					TelegramInfo: func() *domain.TelegramInfo {
						value := p.TelegramInfo.Data()
						return (*domain.TelegramInfo)(&value)
					}(),
					StorageInfo: func() *domain.StorageInfo {
						if p.StorageInfo.Data() == (StorageInfo{}) {
							return nil
						}
						value := p.StorageInfo.Data()
						return &domain.StorageInfo{
							Original: func() *domain.StorageDetail {
								if value.Original == nil {
									return nil
								}
								v := domain.StorageDetail(*value.Original)
								return &v
							}(),
							Regular: func() *domain.StorageDetail {
								if value.Regular == nil {
									return nil
								}
								v := domain.StorageDetail(*value.Regular)
								return &v
							}(),
							Thumb: func() *domain.StorageDetail {
								if value.Thumb == nil {
									return nil
								}
								v := domain.StorageDetail(*value.Thumb)
								return &v
							}(),
						}
					}(),
				})
			}
			return pics
		}(),
		TagIDs: func() *domain.TagIDs {
			if a.Tags == nil {
				return nil
			}
			ids := make([]objectuuid.ObjectUUID, 0, len(a.Tags))
			for _, t := range a.Tags {
				ids = append(ids, t.ID)
			}
			return domain.NewTagIDs(ids...)
		}(),
	}
}
