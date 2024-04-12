package entity

import (
	"gorm.io/gorm"
)

type Artwork struct {
	gorm.Model
	Title        string
	Description  string
	SourceType   string
	SourceURL    string `gorm:"unique"`
	Author       string
	Tags         []Tag      `gorm:"many2many:artwork_tags;"`
	R18          bool       `gorm:"default:false"`
	Pictures     []*Picture `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	MediaGroupID string     `gorm:"unique"`
}

type Tag struct {
	gorm.Model
	Name string `gorm:"unique"`
}

type Picture struct {
	gorm.Model
	DirectURL string
	Width     uint
	Height    uint
	Hash      string
	BlurScore float64
	Format    string
	FilePath  string
	MessageID int
}
