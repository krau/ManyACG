package converter

import (
	"github.com/krau/ManyACG/internal/model/dto"
	"github.com/krau/ManyACG/internal/model/entity"
	"github.com/krau/ManyACG/internal/shared"
	"gorm.io/datatypes"
)

func EntityArtworkToDtoEventItem(ent *entity.Artwork) *dto.ArtworkEventItem {
	if ent == nil {
		return nil
	}
	pics := make([]dto.PictureEventItem, 0, len(ent.Pictures))
	for _, pic := range ent.Pictures {
		pics = append(pics, dto.PictureEventItem{
			ID:         pic.ID,
			ArtworkID:  pic.ArtworkID,
			OrderIndex: pic.OrderIndex,
			Thumbnail:  pic.Thumbnail,
			Original:   pic.Original,
			Width:      pic.Width,
			Height:     pic.Height,
			Phash:      pic.Phash,
			ThumbHash:  pic.ThumbHash,
			CreatedAt:  pic.CreatedAt,
			UpdatedAt:  pic.UpdatedAt,
		})
	}
	item := &dto.ArtworkEventItem{
		ID:             ent.ID,
		Title:          ent.Title,
		Description:    ent.Description,
		R18:            ent.R18,
		CreatedAt:      ent.CreatedAt,
		UpdatedAt:      ent.UpdatedAt,
		SourceType:     ent.SourceType,
		SourceURL:      ent.SourceURL,
		ArtistID:       ent.ArtistID,
		ArtistName:     "",
		ArtistUsername: "",
		ArtistUID:      "",
		Tags:           ent.GetTagsWithAlias(),
		Pictures:       pics,
	}
	if ent.Artist != nil {
		item.ArtistName = ent.Artist.Name
		item.ArtistUsername = ent.Artist.Username
		item.ArtistUID = ent.Artist.UID
	}
	return item
}

func DtoArtworkEventItemToSearchDocument(item *dto.ArtworkEventItem) *dto.ArtworkSearchDocument {
	if item == nil {
		return nil
	}
	return &dto.ArtworkSearchDocument{
		ID:          item.ID.Hex(),
		Title:       item.Title,
		Artist:      item.ArtistName,
		Description: item.Description,
		Tags:        item.Tags,
		R18:         item.R18,
	}
}

func DtoFetchedArtworkToEntityCached(art *dto.FetchedArtwork) *entity.CachedArtwork {
	pics := make([]*entity.CachedPicture, len(art.Pictures))
	for i, pic := range art.Pictures {
		pics[i] = &entity.CachedPicture{
			OrderIndex: pic.Index,
			Thumbnail:  pic.Thumbnail,
			Original:   pic.Original,
			Width:      pic.Width,
			Height:     pic.Height,
		}
	}
	ugoira := &entity.CachedUgoiraMetaData{}
	if art.Ugoira != nil {
		ugoira = &entity.CachedUgoiraMetaData{
			UgoiraMetaData: datatypes.NewJSONType(*art.Ugoira),
		}
	}
	ent := &entity.CachedArtwork{
		SourceURL: art.SourceURL,
		Status:    shared.ArtworkStatusCached,
		Artwork: datatypes.NewJSONType(&entity.CachedArtworkData{
			Title:       art.Title,
			Description: art.Description,
			R18:         art.R18,
			Tags:        art.Tags,
			SourceURL:   art.SourceURL,
			SourceType:  art.SourceType,
			Artist: &entity.CachedArtist{
				Name:     art.Artist.Name,
				UID:      art.Artist.UID,
				Type:     art.Artist.Type,
				Username: art.Artist.Username,
			},
			Pictures:   pics,
			UgoiraMeta: ugoira,
			Version:    1,
		}),
	}
	return ent
}
