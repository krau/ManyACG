package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/duke-git/lancet/v2/slice"
	"github.com/krau/ManyACG/common"
	"github.com/krau/ManyACG/dao"
	"github.com/krau/ManyACG/errs"
	"github.com/krau/ManyACG/storage"
	"github.com/krau/ManyACG/types"
	"github.com/mymmrac/telego"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

func GetRandomTags(ctx context.Context, limit int) ([]string, error) {
	tags, err := dao.GetRandomTags(ctx, limit)
	if err != nil {
		return nil, err
	}
	tagNames := make([]string, 0, len(tags))
	for _, tag := range tags {
		tagNames = append(tagNames, tag.Name)
	}
	return tagNames, nil
}

func GetRandomTagModels(ctx context.Context, limit int) ([]*types.TagModel, error) {
	tags, err := dao.GetRandomTags(ctx, limit)
	if err != nil {
		return nil, err
	}
	return tags, nil
}

func GetTagByName(ctx context.Context, name string) (*types.TagModel, error) {
	return dao.GetTagByName(ctx, name)
}

// 为已有 tag 添加别名
//
// 同时检查是否有其他 tag 的 name 为所指定的别名之一, 在添加完成后, 删除这些 tag, 并将其对应的 artwork 指向新的 tag (即传入的tagID)
func AddTagAliasByID(ctx context.Context, tagID primitive.ObjectID, alias ...string) (*types.TagModel, error) {
	tagModel, err := dao.GetTagByID(ctx, tagID)
	if err != nil {
		return nil, err
	}

	session, err := dao.Client.StartSession()
	if err != nil {
		return nil, err
	}
	defer session.EndSession(ctx)

	if tagModel.Alias == nil {
		tagModel.Alias = make([]string, 0)
	}
	tagAlias := slice.Compact(slice.Unique(slice.Concat(tagModel.Alias, alias)))

	for _, alias := range tagAlias {
		if alias == tagModel.Name {
			continue
		}
		// 检查 alias 是否已经作为其他 tag 的别名存在
		aliasTag, err := dao.GetTagByAlias(ctx, alias)
		if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
			return nil, err
		}

		if aliasTag != nil {
			if aliasTag.ID == tagID {
				continue
			}
			// 处理查询到的 tag name 和 alias 相同的情况
			if alias == aliasTag.Name {
				continue
			}
			return nil, fmt.Errorf("%w: '%s' used by '%s'", errs.ErrAliasAlreadyUsed, alias, aliasTag.Name)
		}
	}

	result, err := session.WithTransaction(ctx, func(ctx mongo.SessionContext) (interface{}, error) {
		if err := dao.UpdateTagAliasByID(ctx, tagID, tagAlias); err != nil {
			return nil, err
		}

		for _, alias := range tagAlias {
			if alias == tagModel.Name {
				continue
			}
			aliasTag, err := dao.GetTagByName(ctx, alias)
			if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
				return nil, err
			}
			if aliasTag == nil {
				continue
			}
			// 添加 aliasTag 中原 alias
			if len(aliasTag.Alias) > 0 {
				regetTagModel, err := dao.GetTagByID(ctx, tagID)
				if err != nil {
					return nil, err
				}
				tagAliasWithaliasTagOriginAlias := slice.Unique(slice.Concat(regetTagModel.Alias, aliasTag.Alias))
				if err := dao.UpdateTagAliasByID(ctx, tagID, tagAliasWithaliasTagOriginAlias); err != nil {
					return nil, err
				}
			}

			// 迁移 artwork 中的 tag
			artworkCollection := dao.GetCollection("Artworks")

			filter := bson.M{"tags": bson.M{"$in": []primitive.ObjectID{aliasTag.ID}}}
			cursor, err := artworkCollection.Find(ctx, filter, options.Find().SetProjection(bson.M{"_id": 1}))
			if err != nil {
				return nil, err
			}
			var matchedDocs []bson.M
			if err = cursor.All(ctx, &matchedDocs); err != nil {
				return nil, err
			}

			if len(matchedDocs) > 0 {
				var docIDs []primitive.ObjectID
				for _, doc := range matchedDocs {
					docIDs = append(docIDs, doc["_id"].(primitive.ObjectID))
				}

				_, err = artworkCollection.UpdateMany(ctx,
					bson.M{"_id": bson.M{"$in": docIDs}},
					bson.M{
						"$pull": bson.M{"tags": aliasTag.ID},
					},
				)
				if err != nil {
					return nil, err
				}

				_, err = artworkCollection.UpdateMany(ctx,
					bson.M{"_id": bson.M{"$in": docIDs}},
					bson.M{
						"$addToSet": bson.M{"tags": tagID},
					},
				)
				if err != nil {
					return nil, err
				}
			}

			// 删除别名对应的 tag
			if _, err := dao.DeleteTagByID(ctx, aliasTag.ID); err != nil {
				return nil, err
			}
		}

		tagModel, err := dao.GetTagByID(ctx, tagID)
		if err != nil {
			return nil, err
		}
		return tagModel, nil
	}, options.Transaction().SetReadPreference(readpref.Primary()))
	if err != nil {
		return nil, err
	}
	return result.(*types.TagModel), nil
}

func PredictArtworkTagsByIDAndUpdate(ctx context.Context, artworkID primitive.ObjectID, tg types.TelegramService) error {
	if common.TaggerClient == nil {
		return errors.New("tagger not available")
	}
	recaption := tg != nil && tg.Bot() != nil
	artwork, err := GetArtworkByID(ctx, artworkID)
	if err != nil {
		return err
	}
	predictedTags := make([]string, 0)
	for _, picture := range artwork.Pictures {
		var pictureFile []byte
		if picture.StorageInfo.Regular != nil {
			pictureFile, err = storage.GetFile(ctx, picture.StorageInfo.Regular)
		} else if picture.StorageInfo.Original != nil {
			pictureFile, err = storage.GetFile(ctx, picture.StorageInfo.Original)
		} else {
			pictureFile, err = common.DownloadWithCache(ctx, picture.Original, nil)
		}
		if err != nil {
			return err
		}
		common.Logger.Debugf("predict picture %s", picture.Original)
		result, err := common.TaggerClient.Predict(ctx, pictureFile)
		if err != nil {
			common.Logger.Errorf("predict picture %s error: %s", picture.Original, err)
			continue
		}
		if len(result.PredictedTags) == 0 {
			continue
		}
		predictedTags = slice.Union(predictedTags, result.PredictedTags)
	}
	newTags := slice.Compact(slice.Union(artwork.Tags, predictedTags))
	if err := UpdateArtworkTagsByURL(ctx, artwork.SourceURL, newTags); err != nil {
		return err
	}
	if recaption {
		go func() {
			artwork, err := GetArtworkByID(ctx, artworkID)
			if err != nil {
				common.Logger.Errorf("get artwork %s error: %s", artworkID.Hex(), err)
				return
			}
			if artwork.Pictures[0].TelegramInfo == nil || artwork.Pictures[0].TelegramInfo.MessageID == 0 {
				return
			}
			artworkHTMLCaption := tg.GetArtworkHTMLCaption(artwork)
			if _, err := tg.Bot().EditMessageCaption(ctx, &telego.EditMessageCaptionParams{
				ChatID:    tg.ChannelChatID(),
				MessageID: artwork.Pictures[0].TelegramInfo.MessageID,
				Caption:   artworkHTMLCaption,
				ParseMode: telego.ModeHTML,
			}); err != nil {
				common.Logger.Errorf("edit message caption error: %s", err)
			}
		}()
	}

	return nil
}

type predictArtworkTagsTask struct {
	ArtworkID primitive.ObjectID
	Ctx       context.Context
	Tg        types.TelegramService
}

var predictArtworkTagsTaskChan = make(chan *predictArtworkTagsTask)

func AddPredictArtworkTagTask(ctx context.Context, artworkID primitive.ObjectID, tg types.TelegramService) {
	predictArtworkTagsTaskChan <- &predictArtworkTagsTask{
		ArtworkID: artworkID,
		Ctx:       ctx,
		Tg:        tg,
	}
}

func listenPredictArtworkTagsTask() {
	for task := range predictArtworkTagsTaskChan {
		if err := PredictArtworkTagsByIDAndUpdate(task.Ctx, task.ArtworkID, task.Tg); err != nil {
			common.Logger.Errorf("predict artwork %s tags error: %s", task.ArtworkID.Hex(), err)
		}
	}
}
