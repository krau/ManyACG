package model

import "github.com/krau/ManyACG/types"

func (picture *PictureModel) ToPicture() *types.Picture {
	return &types.Picture{
		ID:           picture.ID.Hex(),
		ArtworkID:    picture.ArtworkID.Hex(),
		Index:        picture.Index,
		Thumbnail:    picture.Thumbnail,
		Original:     picture.Original,
		Width:        picture.Width,
		Height:       picture.Height,
		Hash:         picture.Hash,
		BlurScore:    picture.BlurScore,
		TelegramInfo: picture.TelegramInfo,
		StorageInfo:  picture.StorageInfo,
	}
}

func (artist *ArtistModel) ToArtist() *types.Artist {
	return &types.Artist{
		ID:       artist.ID.Hex(),
		Name:     artist.Name,
		Type:     artist.Type,
		UID:      artist.UID,
		Username: artist.Username,
	}
}
