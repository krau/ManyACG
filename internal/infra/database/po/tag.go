package po

import (
	"github.com/krau/ManyACG/internal/domain/entity/tag"
	"github.com/krau/ManyACG/pkg/objectuuid"
)

type Tag struct {
	ID    objectuuid.ObjectUUID `gorm:"primaryKey;type:uuid" json:"id"`
	Name  string                `gorm:"type:text;not null;uniqueIndex" json:"name"`
	Alias []TagAlias            `gorm:"foreignKey:TagID;constraint:OnDelete:CASCADE" json:"alias"` // one-to-many relation

	// reverse relation via many2many
	Artworks []*Artwork `gorm:"many2many:artwork_tags" json:"artworks"`
}

func (t *Tag) ToDomain() *tag.Tag {
	if t == nil {
		return nil
	}
	aliases := make([]string, 0, len(t.Alias))
	for _, a := range t.Alias {
		aliases = append(aliases, a.Alias)
	}
	return &tag.Tag{
		ID:    t.ID,
		Name:  t.Name,
		Alias: aliases,
	}
}

type TagAlias struct {
	ID    objectuuid.ObjectUUID `gorm:"primaryKey;type:uuid" json:"id"`
	TagID objectuuid.ObjectUUID `gorm:"type:uuid;index" json:"tag_id"`
	Alias string                `gorm:"type:text;not null;index" json:"alias"`
}
