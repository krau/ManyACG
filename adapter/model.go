package adapter

import (
	"ManyACG/dao"
	"ManyACG/model"
	"ManyACG/types"
	"context"
)

func GetArtworkModelTags(ctx context.Context, artworkModel *model.ArtworkModel) ([]string, error) {
	tags := make([]string, len(artworkModel.Tags))
	for i, tagID := range artworkModel.Tags {
		tagModel, err := dao.GetTagByID(ctx, tagID)
		if err != nil {
			return nil, err
		}
		tags[i] = tagModel.Name
	}
	return tags, nil
}

func GetArtworkModelPictures(ctx context.Context, artworkModel *model.ArtworkModel) ([]*types.Picture, error) {
	pictures := make([]*types.Picture, len(artworkModel.Pictures))
	for i, pictureID := range artworkModel.Pictures {
		pictureModel, err := dao.GetPictureByID(ctx, pictureID)
		if err != nil {
			return nil, err
		}
		pictures[i] = pictureModel.ToPicture()
	}
	return pictures, nil
}

func ConvertToArtwork(ctx context.Context, artworkModel *model.ArtworkModel) (*types.Artwork, error) {
	tags, err := GetArtworkModelTags(ctx, artworkModel)
	if err != nil {
		return nil, err
	}
	pictures, err := GetArtworkModelPictures(ctx, artworkModel)
	if err != nil {
		return nil, err
	}
	artistModel, err := dao.GetArtistByID(ctx, artworkModel.ArtistID)
	if err != nil {
		return nil, err
	}
	return &types.Artwork{
		Title:       artworkModel.Title,
		Description: artworkModel.Description,
		R18:         artworkModel.R18,
		CreatedAt:   artworkModel.CreatedAt.Time(),
		SourceType:  artworkModel.SourceType,
		SourceURL:   artworkModel.SourceURL,
		Artist:      artistModel.ToArtist(),
		Tags:        tags,
		Pictures:    pictures,
	}, nil

}
