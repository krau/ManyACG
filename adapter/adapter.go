package adapter

import (
	"context"
	"fmt"
	"html"
	"sync"

	"github.com/krau/ManyACG/common"
	"github.com/krau/ManyACG/config"
	"github.com/krau/ManyACG/dao"
	"github.com/krau/ManyACG/model"
	"github.com/krau/ManyACG/types"

	"github.com/gorilla/feeds"
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

func GetArtworkModelIndexPicture(ctx context.Context, artworkModel *model.ArtworkModel) (*types.Picture, error) {
	pictureModel, err := dao.GetPictureByID(ctx, artworkModel.Pictures[0])
	if err != nil {
		return nil, err
	}
	return pictureModel.ToPicture(), nil
}

func ConvertToArtwork(ctx context.Context, artworkModel *model.ArtworkModel, opts ...*AdapterOption) (*types.Artwork, error) {
	tags := make([]string, 0)
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
			ID:          artworkModel.ID.Hex(),
			Title:       artworkModel.Title,
			Description: artworkModel.Description,
			R18:         artworkModel.R18,
			LikeCount:   artworkModel.LikeCount,
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
	if option.OnlyIndexPicture && !option.LoadPicture {
		indexPicture, err := GetArtworkModelIndexPicture(ctx, artworkModel)
		if err != nil {
			return nil, err
		}
		pictures = []*types.Picture{indexPicture}
	}
	return &types.Artwork{
		ID:          artworkModel.ID.Hex(),
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

func ConvertToFeedItems(ctx context.Context, artworks []*types.Artwork) []*feeds.Item {
	items := make([]*feeds.Item, len(artworks))
	var wg sync.WaitGroup
	for i, artwork := range artworks {
		wg.Add(1)
		go func(i int, artwork *types.Artwork) {
			defer wg.Done()
			item := &feeds.Item{
				Title:       artwork.Title,
				Link:        &feeds.Link{Href: config.Cfg.API.SiteURL + "/artwork/" + artwork.ID},
				Description: artwork.Description,
				Author:      &feeds.Author{Name: artwork.Artist.Name},
				Created:     artwork.CreatedAt,
				Updated:     artwork.CreatedAt,
				Id:          fmt.Sprintf("%s/artwork/%s", config.Cfg.API.SiteURL, artwork.ID),
				Content: fmt.Sprintf(`
        <article>
            <h2>%s</h2>
            <figure>
                <img src="%s" alt="%s" />
            </figure>
            <p>%s</p>
            <p>Artist: %s</p>
            <p>Created: %s</p>
        </article>
    `,
					html.EscapeString(artwork.Title),
					html.EscapeString(common.ApplyPathRule(artwork.Pictures[0].StorageInfo.Regular.Path)),
					html.EscapeString(artwork.Title),
					html.EscapeString(artwork.Description),
					html.EscapeString(artwork.Artist.Name),
					artwork.CreatedAt.Format("2006-01-02 15:04:05")),
			}
			items[i] = item
		}(i, artwork)
	}
	wg.Wait()
	return items
}
