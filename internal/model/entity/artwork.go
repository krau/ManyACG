package entity

import (
	"time"

	"github.com/duke-git/lancet/v2/slice"
	"github.com/krau/ManyACG/internal/shared"
	"github.com/krau/ManyACG/pkg/objectuuid"
	"gorm.io/gorm"
)

var _ shared.ArtworkLike = (*Artwork)(nil)

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

	// https://www.pixiv.help/hc/en-us/articles/235584628-What-are-Ugoira
	// one-to-many ugoira meta, usually only one
	UgoiraMetas []*UgoiraMeta `gorm:"foreignKey:ArtworkID;constraint:OnDelete:CASCADE" json:"ugoira_meta,omitempty"`
}

// GetType implements shared.ArtworkLike.
func (a *Artwork) GetType() shared.SourceType {
	return a.SourceType
}

// GetUgoiraMetas implements shared.UgoiraArtworkLike.
func (a *Artwork) GetUgoiraMetas() []shared.UgoiraMetaLike {
	var metas []shared.UgoiraMetaLike
	for _, m := range a.UgoiraMetas {
		metas = append(metas, m)
	}
	return metas
}

// GetArtist implements ArtworkLike.
func (a *Artwork) GetArtist() shared.ArtistLike {
	return a.Artist
}

// GetDescription implements ArtworkLike.
func (a *Artwork) GetDescription() string {
	return a.Description
}

// GetPictures implements ArtworkLike.
func (a *Artwork) GetPictures() []shared.PictureLike {
	var pictures []shared.PictureLike
	for _, pic := range a.Pictures {
		pictures = append(pictures, pic)
	}
	return pictures
}

// GetSourceURL implements ArtworkLike.
func (a *Artwork) GetSourceURL() string {
	return a.SourceURL
}

// GetTags implements ArtworkLike.
func (a *Artwork) GetTags() []string {
	var tags []string
	for _, tag := range a.Tags {
		tags = append(tags, tag.Name)
	}
	return tags
}

func (a *Artwork) GetTagsWithAlias() []string {
	tags := make([]string, 0)
	for _, tag := range a.Tags {
		tags = append(tags, tag.Name)
		for _, alias := range tag.Alias {
			tags = append(tags, alias.Alias)
		}
	}
	return slice.Unique(tags)
}

// GetTitle implements ArtworkLike.
func (a *Artwork) GetTitle() string {
	return a.Title
}

func (a *Artwork) GetR18() bool {
	return a.R18
}

func (a *Artwork) GetID() string {
	return a.ID.Hex()
}

func (a *Artwork) BeforeCreate(tx *gorm.DB) (err error) {
	if a.ID == objectuuid.Nil {
		a.ID = objectuuid.New()
	}
	return nil
}
