package service

import (
	"ManyACG/dao"
	"ManyACG/types"
	"context"
)

type stats struct {
	TotalPictures int `json:"total_pictures"`
	TotalTags     int `json:"total_tags"`
	TotalArtists  int `json:"total_artists"`
	TotalArtworks int `json:"total_artworks"`
	// LastArtworkUpdate time.Time `json:"last_artwork_update"`
}

func GetDatabaseStats(ctx context.Context) (*stats, error) {
	totalArtworks, err := dao.GetArtworkCount(ctx, types.R18TypeAll)
	if err != nil {
		return nil, err
	}
	totalArtists, err := dao.GetArtistCount(ctx)
	if err != nil {
		return nil, err
	}
	totalPictures, err := dao.GetPictureCount(ctx)
	if err != nil {
		return nil, err
	}
	totalTags, err := dao.GetTagCount(ctx)
	if err != nil {
		return nil, err
	}
	// lastArtworks, err := dao.GetLatestArtwork(ctx, 1)
	// if err != nil {
	// 	return nil, err
	// }
	// lastArtworkUpdate := time.Now()
	// if len(lastArtworks) > 0 {
	// 	lastArtworkUpdate = lastArtworks[0].CreatedAt.Time()
	// }
	return &stats{
		TotalPictures: int(totalPictures),
		TotalTags:     int(totalTags),
		TotalArtists:  int(totalArtists),
		TotalArtworks: int(totalArtworks),
		// LastArtworkUpdate: lastArtworkUpdate,
	}, nil
}
