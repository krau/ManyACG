package service

import (
	"ManyACG/adapter"
	"ManyACG/dao"
	"ManyACG/types"
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// 查询标题包含指定字符串的作品
func QueryArtworksByTitle(ctx context.Context, title string, r18 types.R18Type, limit int) ([]*types.Artwork, error) {
	if title == "" {
		return GetRandomArtworks(ctx, r18, limit)
	}
	artworkModels, err := dao.QueryArtworksByTitle(ctx, title, r18, int64(limit))
	if err != nil {
		return nil, err
	}
	artworks := make([]*types.Artwork, len(artworkModels))
	for i, artworkModel := range artworkModels {
		artworks[i], err = adapter.ConvertToArtwork(ctx, artworkModel)
		if err != nil {
			return nil, err
		}
	}
	return artworks, nil
}

func QueryArtowrksByDescription(ctx context.Context, description string, r18 types.R18Type, limit int) ([]*types.Artwork, error) {
	if description == "" {
		return GetRandomArtworks(ctx, r18, limit)
	}
	artworkModels, err := dao.QueryArtworksByDescription(ctx, description, r18, int64(limit))
	if err != nil {
		return nil, err
	}
	artworks := make([]*types.Artwork, len(artworkModels))
	for i, artworkModel := range artworkModels {
		artworks[i], err = adapter.ConvertToArtwork(ctx, artworkModel)
		if err != nil {
			return nil, err
		}
	}
	return artworks, nil
}

// 有一种很慢的美
func QueryArtworksByArtistName(ctx context.Context, artistName string, r18 types.R18Type, limit int) ([]*types.Artwork, error) {
	artistModels, err := dao.QueryArtistsByName(ctx, artistName)
	if err != nil {
		return nil, err
	}
	artistIDs := make([]primitive.ObjectID, len(artistModels))
	for i, artistModel := range artistModels {
		artistIDs[i] = artistModel.ID
	}
	averageLimit := limit / len(artistIDs)
	artworks := make([]*types.Artwork, 0)
	for _, arartistID := range artistIDs {
		artworkModels, err := dao.GetArtworksByArtistID(ctx, arartistID, r18, int64(averageLimit))
		if err != nil {
			return nil, err
		}
		for _, artworkModel := range artworkModels {
			artwork, err := adapter.ConvertToArtwork(ctx, artworkModel)
			if err != nil {
				return nil, err
			}
			artworks = append(artworks, artwork)
		}
	}
	if len(artworks) > limit {
		artworks = artworks[:limit]
	}
	return artworks, nil
}

func QueryArtworksByArtistUsername(ctx context.Context, artistUsername string, r18 types.R18Type, limit int) ([]*types.Artwork, error) {
	artistModels, err := dao.QueryArtistsByUserName(ctx, artistUsername)
	if err != nil {
		return nil, err
	}
	artistIDs := make([]primitive.ObjectID, len(artistModels))
	for i, artistModel := range artistModels {
		artistIDs[i] = artistModel.ID
	}
	averageLimit := limit / len(artistIDs)
	artworks := make([]*types.Artwork, 0)
	for _, arartistID := range artistIDs {
		artworkModels, err := dao.GetArtworksByArtistID(ctx, arartistID, r18, int64(averageLimit))
		if err != nil {
			return nil, err
		}
		for _, artworkModel := range artworkModels {
			artwork, err := adapter.ConvertToArtwork(ctx, artworkModel)
			if err != nil {
				return nil, err
			}
			artworks = append(artworks, artwork)
		}
	}
	if len(artworks) > limit {
		artworks = artworks[:limit]
	}
	return artworks, nil
}

// 优先级: tag name > artwork title > artwork description > artist name > artist username
func QueryArtworksByTexts(ctx context.Context, texts [][]string, r18 types.R18Type, limit int) ([]*types.Artwork, error) {
	artworks, err := GetArtworksByTags(ctx, texts, r18, limit)
	if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
		return nil, err
	}
	if len(artworks) >= limit {
		return artworks[:limit], nil
	}

	// 如果 tags 没有找到足够的作品, 则将所有关键词以或的关系查询
	args := make([]string, 0)
	for _, text := range texts {
		args = append(args, text...)
	}

	for _, text := range args {
		artworksByTitle, err := QueryArtworksByTitle(ctx, text, r18, limit)
		if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
			return nil, err
		}
		artworks = append(artworks, artworksByTitle...)
		if len(artworks) >= limit {
			return artworks[:limit], nil
		}
	}

	for _, text := range args {
		artworksByDescription, err := QueryArtowrksByDescription(ctx, text, r18, limit)
		if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
			return nil, err
		}
		artworks = append(artworks, artworksByDescription...)
		if len(artworks) >= limit {
			return artworks[:limit], nil
		}
	}

	for _, text := range args {
		artworksByArtistName, err := QueryArtworksByArtistName(ctx, text, r18, limit)
		if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
			return nil, err
		}
		artworks = append(artworks, artworksByArtistName...)
		if len(artworks) >= limit {
			return artworks[:limit], nil
		}
	}

	for _, text := range args {
		artworksByArtistUsername, err := QueryArtworksByArtistUsername(ctx, text, r18, limit)
		if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
			return nil, err
		}
		artworks = append(artworks, artworksByArtistUsername...)
		if len(artworks) >= limit {
			return artworks[:limit], nil
		}
	}
	if len(artworks) == 0 {
		return nil, mongo.ErrNoDocuments
	}
	return artworks, nil
}
