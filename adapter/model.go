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

func ConvertToArtwork(ctx context.Context, artworkModel *model.ArtworkModel, opts ...*AdapterOption) (*types.Artwork, error) {
	var tags []string
	var pictures []*types.Picture
	var artist *types.Artist
	var err error
	if len(opts) == 0 {
		tags, err = GetArtworkModelTags(ctx, artworkModel)
		if err != nil {
			return nil, err
		}
		pictures, err = GetArtworkModelPictures(ctx, artworkModel)
		if err != nil {
			return nil, err
		}
		artistModel, err := dao.GetArtistByID(ctx, artworkModel.ArtistID)
		if err != nil {
			return nil, err
		}
		artist = artistModel.ToArtist()
		return &types.Artwork{
			Title:       artworkModel.Title,
			Description: artworkModel.Description,
			R18:         artworkModel.R18,
			CreatedAt:   artworkModel.CreatedAt.Time(),
			SourceType:  artworkModel.SourceType,
			SourceURL:   artworkModel.SourceURL,
			Artist:      artist,
			Tags:        tags,
			Pictures:    pictures,
		}, nil
	}
	option := MergeOptions(opts...)
	if option.LoadTag {
		tags, err = GetArtworkModelTags(ctx, artworkModel)
		if err != nil {
			return nil, err
		}
	}
	if option.LoadPicture {
		pictures, err = GetArtworkModelPictures(ctx, artworkModel)
		if err != nil {
			return nil, err
		}
	}
	if option.LoadArtist {
		artistModel, err := dao.GetArtistByID(ctx, artworkModel.ArtistID)
		if err != nil {
			return nil, err
		}
		artist = artistModel.ToArtist()
	}
	return &types.Artwork{
		Title:       artworkModel.Title,
		Description: artworkModel.Description,
		R18:         artworkModel.R18,
		CreatedAt:   artworkModel.CreatedAt.Time(),
		SourceType:  artworkModel.SourceType,
		SourceURL:   artworkModel.SourceURL,
		Artist:      artist,
		Tags:        tags,
		Pictures:    pictures,
	}, nil
}

func ConvertToArtworks(ctx context.Context, artworkModels []*model.ArtworkModel, opts ...*AdapterOption) ([]*types.Artwork, error) {
	if len(artworkModels) == 1 {
		artwork, err := ConvertToArtwork(ctx, artworkModels[0])
		if err != nil {
			return nil, err
		}
		return []*types.Artwork{artwork}, nil
	}
	artworks := make([]*types.Artwork, len(artworkModels))
	errCh := make(chan error, len(artworkModels))
	for i, artworkModel := range artworkModels {
		go func(i int, artworkModel *model.ArtworkModel) {
			artwork, err := ConvertToArtwork(ctx, artworkModel, opts...)
			if err != nil {
				errCh <- err
				return
			}
			artworks[i] = artwork
			errCh <- nil
		}(i, artworkModel)
	}
	for range artworkModels {
		if err := <-errCh; err != nil {
			return nil, err
		}
	}
	return artworks, nil
}
