package dao

import (
	"ManyACG-Bot/model/entity"

	"gorm.io/gorm/clause"
)

func CreateArtwork(artwork *entity.Artwork) error {
	return db.Preload("Tags").Preload("Pictures").Where("source_url = ?", artwork.SourceURL).FirstOrCreate(artwork).Error
}

func UpdateArtwork(artwork *entity.Artwork) error {
	return db.Preload(clause.Associations).Save(artwork).Error
}

func GetArtworkByURL(url string) (*entity.Artwork, error) {
	var artwork *entity.Artwork
	err := db.Preload(clause.Associations).Where("source_url = ?", url).First(artwork).Error
	return artwork, err
}

func DeleteArtworkByURL(url string) error {
	return db.Where("source_url = ?", url).Delete(&entity.Artwork{}).Error
}
