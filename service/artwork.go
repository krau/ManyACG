package service

import (
	"github.com/krau/ManyACG/adapter"
	"github.com/krau/ManyACG/common"
	"github.com/krau/ManyACG/dao"
	"github.com/krau/ManyACG/errs"
	"github.com/krau/ManyACG/sources"

	"context"
	"errors"

	"github.com/krau/ManyACG/types"

	"github.com/duke-git/lancet/v2/slice"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

func CreateArtwork(ctx context.Context, artwork *types.Artwork) (*types.Artwork, error) {
	artworkModel, err := dao.GetArtworkByURL(ctx, artwork.SourceURL)
	if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
		return nil, err
	}
	if artworkModel != nil {
		return nil, errs.ErrArtworkAlreadyExist
	}
	if dao.CheckDeletedByURL(ctx, artwork.SourceURL) {
		return nil, errs.ErrArtworkDeleted
	}

	session, err := dao.Client.StartSession()
	if err != nil {
		return nil, err
	}
	defer session.EndSession(ctx)

	result, err := session.WithTransaction(ctx, func(ctx mongo.SessionContext) (interface{}, error) {
		// 创建 Tag
		tagIDs := make([]primitive.ObjectID, len(artwork.Tags))
		for i, tag := range artwork.Tags {
			tagModel, err := dao.GetTagByName(ctx, tag)
			if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
				return nil, err
			}
			if tagModel != nil {
				tagIDs[i] = tagModel.ID
				continue
			}
			tagModel = &types.TagModel{
				Name: tag,
			}
			tagRes, err := dao.CreateTag(ctx, tagModel)
			if err != nil {
				return nil, err
			}
			tagIDs[i] = tagRes.InsertedID.(primitive.ObjectID)
		}

		seen := make(map[string]struct{})
		resultTags := []primitive.ObjectID{}
		for _, tagID := range tagIDs {
			if _, ok := seen[tagID.Hex()]; !ok {
				resultTags = append(resultTags, tagID)
				seen[tagID.Hex()] = struct{}{}
			}
		}
		tagIDs = resultTags

		// 创建 Artist
		artistModel, err := dao.GetArtistByUID(ctx, artwork.Artist.UID, artwork.Artist.Type)
		if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
			return nil, err
		}
		var artistId primitive.ObjectID
		if artistModel != nil {
			artistModel.Name = artwork.Artist.Name
			artistModel.Username = artwork.Artist.Username
			_, err = dao.UpdateArtist(ctx, artistModel)
			if err != nil {
				return nil, err
			}
			artistId = artistModel.ID
		} else {
			artistModel = &types.ArtistModel{
				Type:     artwork.Artist.Type,
				UID:      artwork.Artist.UID,
				Username: artwork.Artist.Username,
				Name:     artwork.Artist.Name,
			}
			res, err := dao.CreateArtist(ctx, artistModel)
			if err != nil {
				return nil, err
			}
			artistId = res.InsertedID.(primitive.ObjectID)
		}

		// 创建 Artwork
		artworkModel = &types.ArtworkModel{
			Title:       artwork.Title,
			Description: artwork.Description,
			R18:         artwork.R18,
			SourceType:  artwork.SourceType,
			SourceURL:   artwork.SourceURL,
			ArtistID:    artistId,
			Tags:        tagIDs,
		}
		res, err := dao.CreateArtwork(ctx, artworkModel)
		if err != nil {
			return nil, err
		}

		// 创建 Picture
		pictureModels := make([]*types.PictureModel, len(artwork.Pictures))
		for i, picture := range artwork.Pictures {
			var pictureID primitive.ObjectID
			if picture.ID != "" {
				pictureID, err = primitive.ObjectIDFromHex(picture.ID)
				if err != nil {
					return nil, err
				}
			} else {
				pictureID = primitive.NewObjectID()
			}
			pictureModel := &types.PictureModel{
				ID:        pictureID,
				Index:     picture.Index,
				ArtworkID: res.InsertedID.(primitive.ObjectID),
				Thumbnail: picture.Thumbnail,
				Original:  picture.Original,
				Width:     picture.Width,
				Height:    picture.Height,
				Hash:      picture.Hash,
				// BlurScore:    picture.BlurScore,
				TelegramInfo: picture.TelegramInfo,
				StorageInfo:  picture.StorageInfo,
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
	}, options.Transaction().SetReadPreference(readpref.Primary()))
	if err != nil {
		return nil, err
	}
	artworkModel = result.(*types.ArtworkModel)
	return adapter.ConvertToArtwork(ctx, artworkModel)
}

func GetArtworkByURL(ctx context.Context, sourceURL string, opts ...*types.AdapterOption) (*types.Artwork, error) {
	artworkModel, err := dao.GetArtworkByURL(ctx, sourceURL)
	if err != nil {
		return nil, err
	}
	return adapter.ConvertToArtwork(ctx, artworkModel, opts...)
}

// Deprecated: MessageID 现在可能为 0
func GetArtworkByMessageID(ctx context.Context, messageID int, opts ...*types.AdapterOption) (*types.Artwork, error) {
	pictureModel, err := dao.GetPictureByMessageID(ctx, messageID)
	if err != nil {
		return nil, err
	}
	artworkModel, err := dao.GetArtworkByID(ctx, pictureModel.ArtworkID)
	if err != nil {
		return nil, err
	}
	return adapter.ConvertToArtwork(ctx, artworkModel, opts...)
}

func GetArtworkByID(ctx context.Context, id primitive.ObjectID, opts ...*types.AdapterOption) (*types.Artwork, error) {
	artworkModel, err := dao.GetArtworkByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return adapter.ConvertToArtwork(ctx, artworkModel, opts...)
}

func GetArtworkIDByPicture(ctx context.Context, picture *types.Picture) (primitive.ObjectID, error) {
	pictureModel, err := dao.GetPictureByOriginal(ctx, picture.Original)
	if err != nil {
		return primitive.NilObjectID, err
	}
	return pictureModel.ArtworkID, nil
}

func GetRandomArtworks(ctx context.Context, r18 types.R18Type, limit int, convertOpts ...*types.AdapterOption) ([]*types.Artwork, error) {
	artworkModels, err := dao.GetArtworksByR18(ctx, r18, limit)
	if err != nil {
		return nil, err
	}
	artworks, err := adapter.ConvertToArtworks(ctx, artworkModels, convertOpts...)
	if err != nil {
		return nil, err
	}
	return artworks, nil
}

func GetLatestArtworks(ctx context.Context, r18 types.R18Type, page, pageSize int64, convertOpts ...*types.AdapterOption) ([]*types.Artwork, error) {
	artworkModels, err := dao.GetLatestArtworks(ctx, r18, page, pageSize)
	if err != nil {
		return nil, err
	}
	return adapter.ConvertToArtworks(ctx, artworkModels, convertOpts...)
}

// 通过标签获取作品, 标签名使用全字匹配
//
// tags: 二维数组, tags = [["tag1", "tag2"], ["tag3", "tag4"]] 表示 (tag1 || tag2) && (tag3 || tag4)
func GetArtworksByTags(ctx context.Context, tags [][]string, r18 types.R18Type, page, pageSize int64, convertOpts ...*types.AdapterOption) ([]*types.Artwork, error) {
	if len(tags) == 0 {
		return GetRandomArtworks(ctx, r18, int(pageSize))
	}
	tagIDs := make([][]primitive.ObjectID, len(tags))
	for i, tagGroup := range tags {
		tagIDs[i] = make([]primitive.ObjectID, len(tagGroup))
		for j, tagName := range tagGroup {
			tagModel, err := dao.GetTagByName(ctx, tagName)
			if err != nil {
				return nil, err
			}
			tagIDs[i][j] = tagModel.ID
		}
	}
	artworkModels, err := dao.GetArtworksByTags(ctx, tagIDs, r18, page, pageSize)
	if err != nil {
		return nil, err
	}
	return adapter.ConvertToArtworks(ctx, artworkModels, convertOpts...)
}

func GetArtworkCount(ctx context.Context, r18 types.R18Type) (int64, error) {
	return dao.GetArtworkCount(ctx, r18)
}

func GetArtworksByArtistID(ctx context.Context, artistID primitive.ObjectID, r18 types.R18Type, page, pageSize int64, convertOpts ...*types.AdapterOption) ([]*types.Artwork, error) {
	artworkModels, err := dao.GetArtworksByArtistID(ctx, artistID, r18, page, pageSize)
	if err != nil {
		return nil, err
	}
	return adapter.ConvertToArtworks(ctx, artworkModels, convertOpts...)
}

// 使用tag名, 标题, 描述, 作者名, 作者用户名 综合查询
//
// 对于每个关键词, 只要tag名, 标题, 描述, 作者名, 作者用户名中有一个匹配即认为匹配成功
//
// 关键词二维数组中, 每个一维数组中的关键词之间是或的关系, 不同一维数组中的关键词之间是与的关系
func QueryArtworksByTexts(ctx context.Context, texts [][]string, r18 types.R18Type, limit int, convertOpts ...*types.AdapterOption) ([]*types.Artwork, error) {
	artworkModels, err := dao.QueryArtworksByTexts(ctx, texts, r18, limit)
	if err != nil {
		return nil, err
	}
	artworks, err := adapter.ConvertToArtworks(ctx, artworkModels, convertOpts...)
	if err != nil {
		return nil, err
	}
	return artworks, nil
}

func QueryArtworksByTextsPage(ctx context.Context, texts [][]string, r18 types.R18Type, page, pageSize int64, convertOpts ...*types.AdapterOption) ([]*types.Artwork, error) {
	artworkModels, err := dao.QueryArtworksByTextsPage(ctx, texts, r18, page, pageSize)
	if err != nil {
		return nil, err
	}
	artworks, err := adapter.ConvertToArtworks(ctx, artworkModels, convertOpts...)
	if err != nil {
		return nil, err
	}
	return artworks, nil
}

func UpdateArtworkR18ByURL(ctx context.Context, sourceURL string, r18 bool) error {
	artworkModel, err := dao.GetArtworkByURL(ctx, sourceURL)
	if err != nil {
		return err
	}
	_, err = dao.UpdateArtworkR18ByID(ctx, artworkModel.ID, r18)
	if err != nil {
		return err
	}
	return nil
}

func UpdateArtworkR18ByID(ctx context.Context, id primitive.ObjectID, r18 bool) error {
	_, err := dao.UpdateArtworkR18ByID(ctx, id, r18)
	if err != nil {
		return err
	}
	return nil
}

func UpdateArtworkTagsByURL(ctx context.Context, sourceURL string, tags []string) error {
	artworkModel, err := dao.GetArtworkByURL(ctx, sourceURL)
	if err != nil {
		return err
	}

	session, err := dao.Client.StartSession()
	if err != nil {
		return err
	}
	defer session.EndSession(ctx)
	tags = slice.Unique(tags)
	tagIDs := make([]primitive.ObjectID, len(tags))
	_, err = session.WithTransaction(ctx, func(ctx mongo.SessionContext) (interface{}, error) {
		for i, tag := range tags {
			tagModel, err := dao.GetTagByName(ctx, tag)
			if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
				return nil, err
			}
			if tagModel != nil {
				tagIDs[i] = tagModel.ID
				continue
			}
			tagModel = &types.TagModel{
				Name: tag,
			}
			res, err := dao.CreateTag(ctx, tagModel)
			if err != nil {
				return nil, err
			}
			tagIDs[i] = res.InsertedID.(primitive.ObjectID)
		}
		tagIDs = slice.Unique(tagIDs)
		_, err = dao.UpdateArtworkTagsByID(ctx, artworkModel.ID, tagIDs)
		if err != nil {
			return nil, err
		}
		return nil, nil
	}, options.Transaction().SetReadPreference(readpref.Primary()))
	if err != nil {
		return err
	}
	return nil
}

func UpdateArtworkTitleByURL(ctx context.Context, sourceURL, title string) error {
	artworkModel, err := dao.GetArtworkByURL(ctx, sourceURL)
	if err != nil {
		return err
	}
	_, err = dao.UpdateArtworkTitleByID(ctx, artworkModel.ID, title)
	if err != nil {
		return err
	}
	return nil
}

func deleteArtwork(ctx context.Context, id primitive.ObjectID, sourceURL string) error {
	session, err := dao.Client.StartSession()
	if err != nil {
		return err
	}
	defer session.EndSession(ctx)

	_, err = session.WithTransaction(ctx, func(ctx mongo.SessionContext) (interface{}, error) {
		_, err := dao.DeleteArtworkByID(ctx, id)
		if err != nil {
			return nil, err
		}
		_, err = dao.DeletePicturesByArtworkID(ctx, id)
		if err != nil {
			return nil, err
		}
		_, err = dao.CreateDeleted(ctx, &types.DeletedModel{
			SourceURL: sourceURL,
			ArtworkID: id,
		})
		if err != nil {
			return nil, err
		}
		return nil, nil
	}, options.Transaction().SetReadPreference(readpref.Primary()))
	if err != nil {
		return err
	}

	if err := UpdateCachedArtworkStatusByURL(ctx, sourceURL, types.ArtworkStatusCached); err != nil {
		common.Logger.Warnf("更新缓存作品状态失败: %s", err)
	}

	return nil
}

func DeleteArtworkByURL(ctx context.Context, sourceURL string) error {
	artworkModel, err := dao.GetArtworkByURL(ctx, sourceURL)
	if err != nil {
		return err
	}
	return deleteArtwork(ctx, artworkModel.ID, sourceURL)
}

func DeleteArtworkByID(ctx context.Context, id primitive.ObjectID) error {
	artworkModel, err := dao.GetArtworkByID(ctx, id)
	if err != nil {
		return err
	}
	return deleteArtwork(ctx, id, artworkModel.SourceURL)
}

// 用于删除图片后重整 artwork 的 picture 的 index
func TidyArtworkPictureIndexByID(ctx context.Context, artworkID primitive.ObjectID) error {
	artworkModel, err := dao.GetArtworkByID(ctx, artworkID)
	if err != nil {
		return err
	}
	session, err := dao.Client.StartSession()
	if err != nil {
		return err
	}
	defer session.EndSession(ctx)
	_, err = session.WithTransaction(ctx, func(ctx mongo.SessionContext) (any, error) {
		for i, pictureID := range artworkModel.Pictures {
			if _, err := dao.UpdatePictureIndexByID(ctx, pictureID, uint(i)); err != nil {
				return nil, err
			}
		}
		return nil, nil
	}, options.Transaction().SetReadPreference(readpref.Primary()))
	return err
}

func GetArtworkByURLWithCacheFetch(ctx context.Context, sourceURL string) (*types.Artwork, error) {
	artwork, _ := GetArtworkByURL(ctx, sourceURL)
	if artwork != nil {
		return artwork, nil
	}
	cachedArtwork, _ := GetCachedArtworkByURL(ctx, sourceURL)
	if cachedArtwork != nil && cachedArtwork.Artwork != nil {
		return cachedArtwork.Artwork, nil
	}
	artwork, err := sources.GetArtworkInfo(sourceURL)
	if err != nil {
		common.Logger.Errorf("获取作品信息失败: %s", err)
		return nil, errs.ErrFailedToGetArtwork
	}
	err = CreateCachedArtwork(ctx, artwork, types.ArtworkStatusCached)
	if err != nil {
		common.Logger.Warnf("创建缓存作品失败: %s", err)
	}
	return artwork, nil
}
