package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/duke-git/lancet/v2/maputil"
	"github.com/duke-git/lancet/v2/slice"
	"github.com/krau/ManyACG/internal/common/httpclient"
	"github.com/krau/ManyACG/internal/model/entity"
	"github.com/krau/ManyACG/internal/repo"
	"github.com/krau/ManyACG/internal/shared/errs"
	"github.com/krau/ManyACG/pkg/objectuuid"
)

func (s *Service) GetTagByNameWithAlias(ctx context.Context, name string) (*entity.Tag, error) {
	return s.repos.Tag().GetTagByNameWithAlias(ctx, name)
}

func (s *Service) GetTagByName(ctx context.Context, name string) (*entity.Tag, error) {
	return s.repos.Tag().GetTagByName(ctx, name)
}

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

			affectedArtworks, err := repos.Tag().MigrateTagAlias(ctx, aliasTag.ID, tagID)
			if err != nil {
				return err
			}
			if len(affectedArtworks) == 0 {
				continue
			}
			arts, err := repos.Artwork().GetArtworksByIDs(ctx, affectedArtworks)
			if err != nil {
				return err
			}
			for _, art := range arts {
				newTags := make([]*entity.Tag, 0, len(art.Tags)+1)
				exists := false
				for _, t := range art.Tags {
					if t.ID == aliasTag.ID {
						continue
					}
					if t.ID == tag.ID {
						exists = true
					}
					newTags = append(newTags, t)
				}
				if !exists {
					newTags = append(newTags, &entity.Tag{ID: tag.ID, Name: tag.Name})
				}
				if err := repos.Artwork().UpdateArtworkTags(ctx, art.ID, newTags); err != nil {
					return err
				}
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return s.repos.Tag().GetTagByID(ctx, tagID)
}

func (s *Service) PredictAndUpdateArtworkTags(ctx context.Context, artworkID objectuuid.ObjectUUID) error {
	if s.tagger == nil {
		return errs.ErrTaggingNotEnabled
	}
	artwork, err := s.repos.Artwork().GetArtworkByID(ctx, artworkID)
	if err != nil {
		return err
	}
	origTags := slice.Map(artwork.Tags, func(index int, item *entity.Tag) string {
		return item.Name
	})
	predictedTags := make([]string, 0)
	for _, picture := range artwork.Pictures {
		err = func() error {
			if detail := picture.StorageInfo.Data().Original; detail != nil {
				file, err := s.StorageGetFile(ctx, *detail)
				if err != nil {
					return err
				}
				defer file.Close()
				result, err := s.tagger.Predict(ctx, file)
				if err != nil {
					return err
				}
				predictedTags = slice.Union(predictedTags, maputil.Keys(result))
				return nil
			}
			file, clean, err := httpclient.DownloadWithCache(ctx, picture.Original, nil)
			if err != nil {
				return err
			}
			defer clean()
			defer file.Close()
			result, err := s.tagger.Predict(ctx, file)
			if err != nil {
				return err
			}
			predictedTags = slice.Union(predictedTags, maputil.Keys(result))
			return nil
		}()
		if err != nil {
			return err
		}
	}
	newTags := slice.Compact(slice.Union(origTags, predictedTags))
	return s.UpdateArtworkTagsByURL(ctx, artwork.SourceURL, newTags)
}
