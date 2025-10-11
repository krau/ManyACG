package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/duke-git/lancet/v2/slice"
	"github.com/krau/ManyACG/internal/infra/search"
	"github.com/krau/ManyACG/internal/model/command"
	"github.com/krau/ManyACG/internal/model/entity"
	"github.com/krau/ManyACG/internal/model/query"
	"github.com/krau/ManyACG/internal/repo"
	"github.com/krau/ManyACG/internal/shared"
	"github.com/krau/ManyACG/internal/shared/errs"
	"github.com/krau/ManyACG/pkg/objectuuid"
	"gorm.io/datatypes"
)

func (s *Service) CreateArtwork(ctx context.Context, cmd *command.ArtworkCreation) (*entity.Artwork, error) {
	awExist, err := s.repos.Artwork().GetArtworkByURL(ctx, cmd.SourceURL)
	if err != nil && !errors.Is(err, errs.ErrRecordNotFound) {
		return nil, err
	}
	if awExist != nil {
		return nil, errs.ErrArtworkAlreadyExist
	}
	if s.repos.DeletedRecord().CheckDeletedByURL(ctx, cmd.SourceURL) {
		return nil, errs.ErrArtworkDeleted
	}
	err = s.repos.Transaction(ctx, func(repos repo.Repositories) error {
		// 创建 artist
		atsEnt, err := repos.Artist().GetArtistByUID(ctx, cmd.Artist.UID, cmd.SourceType)
		if err != nil && !errors.Is(err, errs.ErrRecordNotFound) {
			return err
		}
		var artistID objectuuid.ObjectUUID
		if atsEnt != nil {
			atsEnt.Name = cmd.Artist.Name
			atsEnt.Username = cmd.Artist.Username
			err := repos.Artist().UpdateArtist(ctx, atsEnt)
			if err != nil {
				return err
			}
			artistID = atsEnt.ID
		} else {
			atsEnt = &entity.Artist{
				Type:     cmd.SourceType,
				UID:      cmd.Artist.UID,
				Username: cmd.Artist.Username,
				Name:     cmd.Artist.Name,
			}
			res, err := repos.Artist().CreateArtist(ctx, atsEnt)
			if err != nil {
				return err
			}
			artistID = *res
		}
		// 创建 tags
		tagEnts := make([]*entity.Tag, 0, len(cmd.Tags))
		tagsStr := slice.Unique(cmd.Tags)
		for _, tag := range tagsStr {
			tagEnt, err := repos.Tag().GetTagByNameWithAlias(ctx, tag)
			if err != nil && !errors.Is(err, errs.ErrRecordNotFound) {
				return err
			}
			if tagEnt != nil {
				if tagEnt.Alias == nil {
					tagEnt.Alias = []entity.TagAlias{}
				}
				tagEnts = append(tagEnts, tagEnt)
				continue
			}
			tagEnt = &entity.Tag{
				Name:  tag,
				Alias: []entity.TagAlias{},
			}
			res, err := repos.Tag().CreateTag(ctx, tagEnt)
			if err != nil {
				return err
			}
			tagEnts = append(tagEnts, res)
		}
		// 创建 artwork
		awEnt := &entity.Artwork{
			Title:       cmd.Title,
			Description: cmd.Description,
			R18:         cmd.R18,
			SourceType:  cmd.SourceType,
			SourceURL:   cmd.SourceURL,
			ArtistID:    artistID,
			Tags:        tagEnts,
		}
		pics := make([]*entity.Picture, 0, len(cmd.Pictures))
		for _, pic := range cmd.Pictures {
			pics = append(pics, &entity.Picture{
				OrderIndex:   pic.Index,
				ArtworkID:    awEnt.ID,
				Thumbnail:    pic.Thumbnail,
				Original:     pic.Original,
				Width:        pic.Width,
				Height:       pic.Height,
				Phash:        pic.Phash,
				ThumbHash:    pic.ThumbHash,
				TelegramInfo: datatypes.NewJSONType(pic.TelegramInfo),
				StorageInfo:  datatypes.NewJSONType(pic.StorageInfo),
			})
		}
		awEnt.Pictures = pics
		if len(cmd.UgoiraMetas) > 0 {
			for _, ugoira := range cmd.UgoiraMetas {
				ugoiraData := &entity.UgoiraMeta{
					OrderIndex:      ugoira.Index,
					Data:            datatypes.NewJSONType(ugoira.Data),
					OriginalStorage: datatypes.NewJSONType(ugoira.OriginalStorage),
					TelegramInfo:    datatypes.NewJSONType(ugoira.TelegramInfo),
				}
				awEnt.UgoiraMetas = append(awEnt.UgoiraMetas, ugoiraData)
			}
		}

		_, err = repos.Artwork().CreateArtwork(ctx, awEnt)
		if err != nil {
			return err
		}
		// update cached artwork status
		cached, err := repos.CachedArtwork().GetCachedArtworkByURL(ctx, cmd.SourceURL)
		if err != nil && !errors.Is(err, errs.ErrRecordNotFound) {
			return err
		}
		if cached != nil {
			cached.Status = shared.ArtworkStatusPosted
			_, err = repos.CachedArtwork().SaveCachedArtwork(ctx, cached)
		}
		return err
	})
	if err != nil {
		return nil, err
	}

	created, err := s.repos.Artwork().GetArtworkByURL(ctx, cmd.SourceURL)
	if err != nil {
		return nil, err
	}

	return created, nil
}

func (s *Service) UpdateArtworkR18ByURL(ctx context.Context, sourceURL string, r18 bool) error {
	awEnt, err := s.repos.Artwork().GetArtworkByURL(ctx, sourceURL)
	if err != nil {
		return err
	}
	return s.repos.Artwork().UpdateArtworkByMap(ctx, awEnt.ID, map[string]any{"r18": r18})
}

func (s *Service) UpdateArtworkR18ByID(ctx context.Context, id objectuuid.ObjectUUID, r18 bool) error {
	return s.repos.Artwork().UpdateArtworkByMap(ctx, id, map[string]any{"r18": r18})
}

func (s *Service) UpdateArtworkTagsByURL(ctx context.Context, sourceURL string, tags []string) error {
	awEnt, err := s.repos.Artwork().GetArtworkByURL(ctx, sourceURL)
	if err != nil {
		return err
	}
	uniTags := slice.Unique(tags)
	return s.repos.Transaction(ctx, func(repos repo.Repositories) error {
		tagEnts := make([]*entity.Tag, 0, len(uniTags))
		for _, tag := range uniTags {
			tagEnt, err := repos.Tag().GetTagByNameWithAlias(ctx, tag)
			if err != nil && !errors.Is(err, errs.ErrRecordNotFound) {
				return err
			}
			if tagEnt != nil {
				tagEnts = append(tagEnts, tagEnt)
				continue
			}
			tagEnt = &entity.Tag{
				Name: tag,
			}
			res, err := repos.Tag().CreateTag(ctx, tagEnt)
			if err != nil {
				return err
			}
			tagEnts = append(tagEnts, res)
			// [TODO] tag 排序
		}
		return repos.Artwork().UpdateArtworkTags(ctx, awEnt.ID, tagEnts)
	})
}

func (s *Service) GetArtworkByURL(ctx context.Context, sourceURL string) (*entity.Artwork, error) {
	return s.repos.Artwork().GetArtworkByURL(ctx, sourceURL)
}

func (s *Service) DeleteArtworkByURL(ctx context.Context, sourceURL string) error {
	awEnt, err := s.repos.Artwork().GetArtworkByURL(ctx, sourceURL)
	if err != nil {
		return err
	}
	err = s.repos.Transaction(ctx, func(repos repo.Repositories) error {
		err := repos.DeletedRecord().CreateDeletedRecord(ctx, &entity.DeletedRecord{
			SourceURL: sourceURL,
		})
		if err != nil {
			return err
		}
		return repos.Artwork().DeleteArtworkByID(ctx, awEnt.ID)
	})
	return err
}

func (s *Service) GetArtworkByID(ctx context.Context, id objectuuid.ObjectUUID) (*entity.Artwork, error) {
	return s.repos.Artwork().GetArtworkByID(ctx, id)
}

func (s *Service) DeleteArtworkByID(ctx context.Context, id objectuuid.ObjectUUID) error {
	awEnt, err := s.repos.Artwork().GetArtworkByID(ctx, id)
	if err != nil {
		return err
	}
	err = s.repos.Transaction(ctx, func(repos repo.Repositories) error {
		err := repos.DeletedRecord().CreateDeletedRecord(ctx, &entity.DeletedRecord{
			SourceURL: awEnt.SourceURL,
		})
		if err != nil {
			return err
		}
		return repos.Artwork().DeleteArtworkByID(ctx, id)
	})
	return err
}

func (s *Service) UpdateArtworkTitleByURL(ctx context.Context, sourceURL, title string) error {
	awEnt, err := s.repos.Artwork().GetArtworkByURL(ctx, sourceURL)
	if err != nil {
		return err
	}
	return s.repos.Artwork().UpdateArtworkByMap(ctx, awEnt.ID, map[string]any{"title": title})
}

func (s *Service) QueryArtworks(ctx context.Context, que query.ArtworksDB) ([]*entity.Artwork, error) {
	return s.repos.Artwork().QueryArtworks(ctx, que)
}

func (s *Service) FindSimilarArtworks(ctx context.Context, que *query.ArtworkSimilar) ([]*entity.Artwork, error) {
	if s.searcher == nil {
		return nil, search.ErrNotEnabled
	}
	result, err := s.searcher.FindSimilarArtworks(ctx, que)
	if err != nil {
		return nil, fmt.Errorf("find similar artworks failed: %w", err)
	}
	if len(result.IDs) == 0 {
		return []*entity.Artwork{}, nil
	}
	artworks, err := s.repos.Artwork().GetArtworksByIDs(ctx, result.IDs)
	if err != nil {
		return nil, fmt.Errorf("get artworks by ids failed: %w", err)
	}
	return artworks, nil
}

func (s *Service) SearchArtworks(ctx context.Context, que *query.ArtworkSearch) ([]*entity.Artwork, error) {
	if s.searcher == nil {
		return nil, search.ErrNotEnabled
	}
	result, err := s.searcher.SearchArtworks(ctx, que)
	if err != nil {
		return nil, fmt.Errorf("search artworks failed: %w", err)
	}
	if len(result.IDs) == 0 {
		return []*entity.Artwork{}, nil
	}
	artworks, err := s.repos.Artwork().GetArtworksByIDs(ctx, result.IDs)
	if err != nil {
		return nil, fmt.Errorf("get artworks by ids failed: %w", err)
	}
	return artworks, nil
}
