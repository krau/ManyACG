package service

// func CreateCachedArtwork(ctx context.Context, artwork *types.Artwork, status types.ArtworkStatus) error {
// 	// _, err := dao.CreateCachedArtwork(ctx, artwork, status)
// 	// return err
// 	data := &entity.CachedArtworkData{
// 		ID:          artwork.ID,
// 		Title:       artwork.Title,
// 		Description: artwork.Description,
// 		SourceType:  shared.SourceType(artwork.SourceType),
// 		SourceURL:   artwork.SourceURL,
// 		R18:         artwork.R18,
// 		Artist: &entity.CachedArtist{
// 			Name:     artwork.Artist.Name,
// 			UID:      artwork.Artist.UID,
// 			Type:     shared.SourceType(artwork.Artist.Type),
// 			Username: artwork.Artist.Username,
// 			ID:       artwork.Artist.ID,
// 		},
// 		Tags: artwork.Tags,
// 		Pictures: func() []*entity.CachedPicture {
// 			pictures := make([]*entity.CachedPicture, len(artwork.Pictures))
// 			for i, pic := range artwork.Pictures {
// 				pictures[i] = &entity.CachedPicture{
// 					ID:        pic.ID,
// 					ArtworkID: pic.ArtworkID,
// 					Index:     pic.Index,
// 					Thumbnail: pic.Thumbnail,
// 					Original:  pic.Original,
// 					Width:     pic.Width,
// 					Height:    pic.Height,
// 					Phash:     pic.Hash,
// 					ThumbHash: pic.ThumbHash,
// 				}
// 			}
// 			return pictures
// 		}(),
// 	}
// 	cacheModel := &entity.CachedArtwork{
// 		SourceURL: artwork.SourceURL,
// 		Artwork:   datatypes.NewJSONType(data),
// 		Status:    shared.ArtworkStatus(status),
// 	}
// 	if _, err := database.Default().CreateCachedArtwork(ctx, cacheModel); err != nil {
// 		return err
// 	}
// 	return nil
// }

// func GetCachedArtworkByURL(ctx context.Context, sourceURL string) (*types.CachedArtworksModel, error) {
// 	cachedArtwork, err := dao.GetCachedArtworkByURL(ctx, sourceURL)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return cachedArtwork, nil
// 	// cachedArtwork, err := database.Default().GetCachedArtworkByURL(ctx, sourceURL)
// 	// if err != nil {
// 	// 	return nil, err
// 	// }
// 	// return cachedArtwork, nil
// }

// func UpdateCachedArtworkStatusByURL(ctx context.Context, sourceURL string, status types.ArtworkStatus) error {
// 	_, err := dao.UpdateCachedArtworkStatusByURL(ctx, sourceURL, status)
// 	return err
// }

// func UpdateCachedArtwork(ctx context.Context, artwork *types.CachedArtworksModel) error {
// 	_, err := dao.UpdateCachedArtwork(ctx, artwork)
// 	return err
// }

// GetCachedArtworkByURLWithCache get cached artwork by sourceURL, if not exist, fetch from source and cache it
// func GetCachedArtworkByURLWithCache(ctx context.Context, sourceURL string) (*types.CachedArtworksModel, error) {
// 	cachedArtwork, err := dao.GetCachedArtworkByURL(ctx, sourceURL)
// 	if err != nil {
// 		artwork, err := sources.GetArtworkInfo(sourceURL)
// 		if err != nil {
// 			return nil, err
// 		}
// 		err = CreateCachedArtwork(ctx, artwork, types.ArtworkStatusCached)
// 		if err != nil {
// 			return nil, err
// 		}
// 		cachedArtwork, err = dao.GetCachedArtworkByURL(ctx, artwork.SourceURL)
// 		if err != nil {
// 			return nil, err
// 		}
// 	}
// 	return cachedArtwork, nil
// }

// func DeleteCachedArtworkByURL(ctx context.Context, sourceURL string) error {
// 	_, err := dao.DeleteCachedArtworkByURL(ctx, sourceURL)
// 	return err
// }

// func DeleteCachedArtworkPicture(ctx context.Context, cachedArtwork *types.CachedArtworksModel, pictureIndex int) error {
// 	if pictureIndex < 0 || pictureIndex > len(cachedArtwork.Artwork.Pictures) {
// 		return errs.ErrIndexOOB
// 	}
// 	cachedArtwork.Artwork.Pictures = append(cachedArtwork.Artwork.Pictures[:pictureIndex], cachedArtwork.Artwork.Pictures[pictureIndex+1:]...)
// 	for i := pictureIndex; i < len(cachedArtwork.Artwork.Pictures); i++ {
// 		cachedArtwork.Artwork.Pictures[i].Index = uint(i)
// 	}
// 	err := UpdateCachedArtwork(ctx, cachedArtwork)
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }

