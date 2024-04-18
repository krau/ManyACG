package service

import (
	"ManyACG-Bot/dao"
	"ManyACG-Bot/dao/model"
	es "ManyACG-Bot/errors"
	"ManyACG-Bot/types"
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func CreateArtwork(ctx context.Context, artwork *types.Artwork) (*types.Artwork, error) {
	artworkModel, err := dao.GetArtworkByURL(ctx, artwork.SourceURL)
	if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
		return nil, err
	}
	if artworkModel != nil {
		return nil, es.ErrArtworkAlreadyExist
	}

	session, err := dao.Client.StartSession()
	if err != nil {
		return nil, err
	}
	defer session.EndSession(ctx)

	result, err := session.WithTransaction(ctx, func(ctx mongo.SessionContext) (interface{}, error) {

		// 创建 Tag
		tagIDs := make([]primitive.ObjectID, len(artwork.Tags))
		newTagCount := 0
		for _, tag := range artwork.Tags {
			tagModel, err := dao.GetTagByName(ctx, tag)
			if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
				return nil, err
			}
			if tagModel != nil {
				tagIDs = append(tagIDs, tagModel.ID)
				continue
			}
			tagModel = &model.TagModel{
				Name: tag,
			}
			tagRes, err := dao.CreateTag(ctx, tagModel)
			if err != nil {
				return nil, err
			}
			tagIDs[newTagCount] = tagRes.InsertedID.(primitive.ObjectID)
			newTagCount++
		}

		// 创建 Artist
		var artist_id primitive.ObjectID
		artistModel, err := dao.GetArtistByUID(ctx, artwork.Artist.UID)
		if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
			return nil, err
		}
		if artistModel != nil {
			artist_id = artistModel.ID
		} else {
			artistModel = &model.ArtistModel{
				Type:     artwork.Artist.Type,
				UID:      artwork.Artist.UID,
				Username: artwork.Artist.Username,
				Name:     artwork.Artist.Name,
			}
			res, err := dao.CreateArtist(ctx, artistModel)
			if err != nil {
				return nil, err
			}
			artist_id = res.InsertedID.(primitive.ObjectID)
		}

		// 创建 Artwork
		artworkModel = &model.ArtworkModel{
			Title:       artwork.Title,
			Description: artwork.Description,
			R18:         artwork.R18,
			SourceType:  artwork.SourceType,
			SourceURL:   artwork.SourceURL,
			ArtistID:    artist_id,
			Tags:        tagIDs,
		}
		res, err := dao.CreateArtwork(ctx, artworkModel)
		if err != nil {
			return nil, err
		}

		// 创建 Picture
		pictureModels := make([]*model.PictureModel, len(artwork.Pictures))
		for i, picture := range artwork.Pictures {
			pictureModel := &model.PictureModel{
				Index:        picture.Index,
				ArtworkID:    res.InsertedID.(primitive.ObjectID),
				Thumbnail:    picture.Thumbnail,
				Original:     picture.Original,
				Width:        picture.Width,
				Height:       picture.Height,
				Hash:         picture.Hash,
				BlurScore:    picture.BlurScore,
				TelegramInfo: (*model.TelegramInfo)(picture.TelegramInfo),
				StorageInfo:  (*model.StorageInfo)(picture.StorageInfo),
			}
			pictureModels[i] = pictureModel
		}
		pictureRes, err := dao.CreatePictures(ctx, pictureModels)
		if err != nil {
			return nil, err
		}
		pictureIDs := make([]primitive.ObjectID, len(pictureRes.InsertedIDs))
		for i, id := range pictureRes.InsertedIDs {
			pictureIDs[i] = id.(primitive.ObjectID)
		}

		// 更新 Artwork 的 Pictures
		_, err = dao.UpdateArtworkPicturesByID(ctx, res.InsertedID.(primitive.ObjectID), pictureIDs)
		if err != nil {
			return nil, err
		}
		artworkModel, err = dao.GetArtworkByID(ctx, res.InsertedID.(primitive.ObjectID))
		if err != nil {
			return nil, err
		}
		return artworkModel, nil
	})
	if err != nil {
		return nil, err
	}
	artwork.CreatedAt = result.(*model.ArtworkModel).CreatedAt.Time()
	return artwork, nil
}
