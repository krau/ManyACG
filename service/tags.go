package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/duke-git/lancet/v2/slice"
	"github.com/krau/ManyACG/internal/model/entity"
	"github.com/krau/ManyACG/internal/repo"
	"github.com/krau/ManyACG/internal/shared/errs"
	"github.com/krau/ManyACG/pkg/objectuuid"
)

func (s *Service) GetTagByNameWithAlias(ctx context.Context, name string) (*entity.Tag, error) {
	tag, err := s.repos.Tag().GetTagByName(ctx, name)
	if err == nil {
		return tag, nil
	}
	alias, err := s.repos.Tag().GetAliasTagByName(ctx, name)
	if err != nil {
		return nil, err
	}
	tag, err = s.repos.Tag().GetTagByID(ctx, alias.TagID)
	if err != nil {
		return nil, err
	}
	return tag, nil
}

// func GetRandomTags(ctx context.Context, limit int) ([]string, error) {
// 	tags, err := dao.GetRandomTags(ctx, limit)
// 	if err != nil {
// 		return nil, err
// 	}
// 	tagNames := make([]string, 0, len(tags))
// 	for _, tag := range tags {
// 		tagNames = append(tagNames, tag.Name)
// 	}
// 	return tagNames, nil
// }

// func GetRandomTagModels(ctx context.Context, limit int) ([]*types.TagModel, error) {
// 	tags, err := dao.GetRandomTags(ctx, limit)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return tags, nil
// }

// func GetTagByName(ctx context.Context, name string) (*types.TagModel, error) {
// 	return dao.GetTagByName(ctx, name)
// }

// 为已有 tag 添加别名
//
// 同时检查是否有其他 tag 的 name 为所指定的别名之一, 在添加完成后, 删除这些 tag, 并将其对应的 artwork 指向新的 tag (即传入的tagID)
// func AddTagAliasByID(ctx context.Context, tagID primitive.ObjectID, alias ...string) (*types.TagModel, error) {
// 	tagModel, err := dao.GetTagByID(ctx, tagID)
// 	if err != nil {
// 		return nil, err
// 	}

// 	session, err := dao.Client.StartSession()
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer session.EndSession(ctx)

// 	if tagModel.Alias == nil {
// 		tagModel.Alias = make([]string, 0)
// 	}
// 	tagAlias := slice.Compact(slice.Unique(slice.Concat(tagModel.Alias, alias)))

// 	for _, alias := range tagAlias {
// 		if alias == tagModel.Name {
// 			continue
// 		}
// 		// 检查 alias 是否已经作为其他 tag 的别名存在
// 		aliasTag, err := dao.GetTagByAlias(ctx, alias)
// 		if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
// 			return nil, err
// 		}

// 		if aliasTag != nil {
// 			if aliasTag.ID == tagID {
// 				continue
// 			}
// 			// 处理查询到的 tag name 和 alias 相同的情况
// 			if alias == aliasTag.Name {
// 				continue
// 			}
// 			return nil, fmt.Errorf("%w: '%s' used by '%s'", errs.ErrAliasAlreadyUsed, alias, aliasTag.Name)
// 		}
// 	}

// 	result, err := session.WithTransaction(ctx, func(ctx mongo.SessionContext) (interface{}, error) {
// 		if err := dao.UpdateTagAliasByID(ctx, tagID, tagAlias); err != nil {
// 			return nil, err
// 		}

// 		for _, alias := range tagAlias {
// 			if alias == tagModel.Name {
// 				continue
// 			}
// 			aliasTag, err := dao.GetTagByName(ctx, alias)
// 			if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
// 				return nil, err
// 			}
// 			if aliasTag == nil {
// 				continue
// 			}
// 			// 添加 aliasTag 中原 alias
// 			if len(aliasTag.Alias) > 0 {
// 				regetTagModel, err := dao.GetTagByID(ctx, tagID)
// 				if err != nil {
// 					return nil, err
// 				}
// 				tagAliasWithaliasTagOriginAlias := slice.Unique(slice.Concat(regetTagModel.Alias, aliasTag.Alias))
// 				if err := dao.UpdateTagAliasByID(ctx, tagID, tagAliasWithaliasTagOriginAlias); err != nil {
// 					return nil, err
// 				}
// 			}

// 			// 迁移 artwork 中的 tag
// 			artworkCollection := dao.GetCollection("Artworks")

// 			filter := bson.M{"tags": bson.M{"$in": []primitive.ObjectID{aliasTag.ID}}}
// 			cursor, err := artworkCollection.Find(ctx, filter, options.Find().SetProjection(bson.M{"_id": 1}))
// 			if err != nil {
// 				return nil, err
// 			}
// 			var matchedDocs []bson.M
// 			if err = cursor.All(ctx, &matchedDocs); err != nil {
// 				return nil, err
// 			}

// 			if len(matchedDocs) > 0 {
// 				var docIDs []primitive.ObjectID
// 				for _, doc := range matchedDocs {
// 					docIDs = append(docIDs, doc["_id"].(primitive.ObjectID))
// 				}

// 				_, err = artworkCollection.UpdateMany(ctx,
// 					bson.M{"_id": bson.M{"$in": docIDs}},
// 					bson.M{
// 						"$pull": bson.M{"tags": aliasTag.ID},
// 					},
// 				)
// 				if err != nil {
// 					return nil, err
// 				}

// 				_, err = artworkCollection.UpdateMany(ctx,
// 					bson.M{"_id": bson.M{"$in": docIDs}},
// 					bson.M{
// 						"$addToSet": bson.M{"tags": tagID},
// 					},
// 				)
// 				if err != nil {
// 					return nil, err
// 				}
// 			}

// 			// 删除别名对应的 tag
// 			if _, err := dao.DeleteTagByID(ctx, aliasTag.ID); err != nil {
// 				return nil, err
// 			}
// 		}

// 		tagModel, err := dao.GetTagByID(ctx, tagID)
// 		if err != nil {
// 			return nil, err
// 		}
// 		return tagModel, nil
// 	}, options.Transaction().SetReadPreference(readpref.Primary()))
// 	if err != nil {
// 		return nil, err
// 	}
// 	return result.(*types.TagModel), nil
// }

// 为已有 tag 添加别名
//
// 同时检查是否有其他 tag 的 name 为所指定的别名之一, 在添加完成后, 删除这些 tag, 并将其对应的 artwork 添加这个新的 tag (即传入的tagID)
func (s *Service) AddTagAlias(ctx context.Context, tagID objectuuid.ObjectUUID, alias []string) (*entity.Tag, error) {
	tag, err := s.repos.Tag().GetTagByID(ctx, tagID)
	if err != nil {
		return nil, err
	}

	uniAlias := slice.Compact(slice.Unique(alias))
	aliasEnts := make([]*entity.TagAlias, 0, len(uniAlias))
	for _, a := range uniAlias {
		aliasEnts = append(aliasEnts, &entity.TagAlias{
			Alias: a,
		})
	}
	for _, a := range aliasEnts {
		existTag, err := s.repos.Tag().GetTagByName(ctx, a.Alias)
		if err != nil && !errors.Is(err, errs.ErrRecordNotFound) {
			return nil, err
		}
		if existTag != nil {
			if existTag.ID == tagID {
				continue
			}
			// 处理查询到的 tag name 和 alias 相同的情况
			if a.Alias == existTag.Name {
				continue
			}
			return nil, fmt.Errorf("%w: '%s' used by '%s'", errs.ErrAliasAlreadyUsed, a.Alias, existTag.Name)
		}
	}
	err = s.repos.Transaction(ctx, func(repos repo.Repositories) error {
		if err := repos.Tag().UpdateTagAlias(ctx, tag.ID, aliasEnts); err != nil {
			return err
		}
		for _, a := range aliasEnts {
			aliasTag, err := repos.Tag().GetTagByName(ctx, a.Alias)
			if err != nil && !errors.Is(err, errs.ErrRecordNotFound) {
				return err
			}
			if aliasTag == nil || aliasTag.ID == tagID {
				continue
			}

			// 如果 aliasTag 还有自己的别名，合并进目标 tag
			if len(aliasTag.Alias) > 0 {
				// 重新拉取目标 tag，避免脏读
				reloadedTag, err := repos.Tag().GetTagByID(ctx, tagID)
				if err != nil {
					return err
				}
				aliasSet := make(map[string]struct{})
				for _, aa := range append(reloadedTag.Alias, aliasTag.Alias...) {
					if aa.Alias == "" || aa.Alias == reloadedTag.Name {
						continue
					}
					aliasSet[aa.Alias] = struct{}{}
				}
				newAliases := make([]*entity.TagAlias, 0, len(aliasSet))
				for aa := range aliasSet {
					newAliases = append(newAliases, &entity.TagAlias{Alias: aa})
				}
				if err := repos.Tag().UpdateTagAlias(ctx, reloadedTag.ID, newAliases); err != nil {
					return err
				}
			}
		
			if err := repos.Tag().MigrateTagAlias(ctx, aliasTag.ID, tag.ID); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return s.repos.Tag().GetTagByID(ctx, tagID)
}

// func PredictArtworkTags(ctx context.Context, artworkID objectuuid.ObjectUUID) error {
// 	if common.TaggerClient == nil {
// 		return errors.New("tagger not available")
// 	}
// 	artwork, err := database.Default().GetArtworkByID(ctx, artworkID)
// 	if err != nil {
// 		return err
// 	}
// 	origTags := slice.Map(artwork.Tags, func(index int, item *entity.Tag) string {
// 		return item.Name
// 	})
// 	predictedTags := make([]string, 0)
// 	for _, picture := range artwork.Pictures {
// 		var pictureFile []byte
// 		if picture.StorageInfo.Data().Regular != nil {
// 			pictureFile, err = storage.GetFile(ctx, picture.StorageInfo.Data().Regular)
// 		} else if picture.StorageInfo.Data().Original != nil {
// 			pictureFile, err = storage.GetFile(ctx, picture.StorageInfo.Data().Original)
// 		} else {
// 			pictureFile, err = common.DownloadWithCache(ctx, picture.Original, nil)
// 		}
// 		if err != nil {
// 			return err
// 		}
// 		common.Logger.Debugf("predict picture %s", picture.Original)
// 		result, err := common.TaggerClient.Predict(ctx, pictureFile)
// 		if err != nil {
// 			common.Logger.Errorf("predict picture %s error: %s", picture.Original, err)
// 			continue
// 		}
// 		if len(result.PredictedTags) == 0 {
// 			continue
// 		}
// 		predictedTags = slice.Union(predictedTags, result.PredictedTags)
// 	}
// 	newTags := slice.Compact(slice.Union(origTags, predictedTags))
// 	if err := UpdateArtworkTagsByURL(ctx, artwork.SourceURL, newTags); err != nil {
// 		return err
// 	}
// 	// if recaption {
// 	// 	go func() {
// 	// 		artwork, err := GetArtworkByID(ctx, artworkID)
// 	// 		if err != nil {
// 	// 			common.Logger.Errorf("get artwork %s error: %s", artworkID.Hex(), err)
// 	// 			return
// 	// 		}
// 	// 		if artwork.Pictures[0].TelegramInfo == nil || artwork.Pictures[0].TelegramInfo.MessageID == 0 {
// 	// 			return
// 	// 		}
// 	// 		artworkHTMLCaption := tg.GetArtworkHTMLCaption(artwork)
// 	// 		if _, err := tg.Bot().EditMessageCaption(ctx, &telego.EditMessageCaptionParams{
// 	// 			ChatID:    tg.ChannelChatID(),
// 	// 			MessageID: artwork.Pictures[0].TelegramInfo.MessageID,
// 	// 			Caption:   artworkHTMLCaption,
// 	// 			ParseMode: telego.ModeHTML,
// 	// 		}); err != nil {
// 	// 			common.Logger.Errorf("edit message caption error: %s", err)
// 	// 		}
// 	// 	}()
// 	// }

// 	return nil
// }

// type predictArtworkTagsTask struct {
// 	ArtworkID objectuuid.ObjectUUID
// 	Ctx       context.Context
// }

// var predictArtworkTagsTaskChan = make(chan *predictArtworkTagsTask)

// func AddPredictArtworkTagTask(ctx context.Context, artworkID objectuuid.ObjectUUID) {
// 	predictArtworkTagsTaskChan <- &predictArtworkTagsTask{
// 		ArtworkID: artworkID,
// 		Ctx:       ctx,
// 	}
// }

// func listenPredictArtworkTagsTask() {
// 	for task := range predictArtworkTagsTaskChan {
// 		if err := PredictArtworkTags(task.Ctx, task.ArtworkID); err != nil {
// 			common.Logger.Errorf("predict artwork %s tags error: %s", task.ArtworkID.Hex(), err)
// 		}
// 	}
// }
