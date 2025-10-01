package service

import (
	"context"
	"errors"

	"github.com/duke-git/lancet/v2/slice"
	"github.com/krau/ManyACG/errs"
	"github.com/krau/ManyACG/internal/infra/database"
	"github.com/krau/ManyACG/internal/model/command"
	"github.com/krau/ManyACG/internal/model/entity"
	"github.com/krau/ManyACG/pkg/objectuuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

func CreateArtwork(ctx context.Context, cmd *command.ArtworkCreation) (*entity.Artwork, error) {
	awExist, err := database.Default().GetArtworkByURL(ctx, cmd.SourceURL)
	if err != nil && !errors.Is(err, database.ErrRecordNotFound) {
		return nil, err
	}
	if awExist != nil {
		return nil, errs.ErrArtworkAlreadyExist
	}
	if database.Default().CheckDeletedByURL(ctx, cmd.SourceURL) {
		return nil, errs.ErrArtworkDeleted
	}
	// 创建 artist
	err = database.Default().Transaction(ctx, func(tx *database.DB, _ *gorm.DB) error {
		atsEnt, err := database.Default().GetArtistByUID(ctx, cmd.Artist.UID, cmd.SourceType)
		if err != nil && !errors.Is(err, database.ErrRecordNotFound) {
			return err
		}
		var artistID objectuuid.ObjectUUID
		if atsEnt != nil {
			atsEnt.Name = cmd.Artist.Name
			atsEnt.Username = cmd.Artist.Username
			err := database.Default().UpdateArtist(ctx, atsEnt)
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
			res, err := database.Default().CreateArtist(ctx, atsEnt)
			if err != nil {
				return err
			}
			artistID = *res
		}
		// 创建 tags
		tagEnts := make([]*entity.Tag, len(cmd.Tags))
		tagsStr := slice.Unique(cmd.Tags)
		for _, tag := range tagsStr {
			tagEnt, err := GetTagByNameWithAlias(ctx, tag)
			if err != nil && !errors.Is(err, database.ErrRecordNotFound) {
				return err
			}
			if tagEnt != nil {
				tagEnts = append(tagEnts, tagEnt)
				continue
			}
			tagEnt = &entity.Tag{
				Name: tag,
			}
			res, err := database.Default().CreateTag(ctx, tagEnt)
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
		pics := make([]*entity.Picture, len(cmd.Pictures))
		for i, pic := range cmd.Pictures {
			pics[i] = &entity.Picture{
				Index:        pic.Index,
				ArtworkID:    awEnt.ID,
				Thumbnail:    pic.Thumbnail,
				Original:     pic.Original,
				Width:        pic.Width,
				Height:       pic.Height,
				Phash:        pic.Phash,
				ThumbHash:    pic.ThumbHash,
				TelegramInfo: datatypes.NewJSONType(*pic.TelegramInfo),
				StorageInfo:  datatypes.NewJSONType(*pic.StorageInfo),
			}
		}
		awEnt.Pictures = pics
		_, err = database.Default().CreateArtwork(ctx, awEnt)
		return err
	})
	if err != nil {
		return nil, err
	}

	created, err := database.Default().GetArtworkByURL(ctx, cmd.SourceURL)
	if err != nil {
		return nil, err
	}

	return created, nil
}

func UpdateArtworkR18ByURL(ctx context.Context, sourceURL string, r18 bool) error {
	awEnt, err := database.Default().GetArtworkByURL(ctx, sourceURL)
	if err != nil {
		return err
	}
	return database.Default().UpdateArtworkByMap(ctx, awEnt.ID, map[string]any{"r18": r18})
}

func UpdateArtworkR18ByID(ctx context.Context, id objectuuid.ObjectUUID, r18 bool) error {
	return database.Default().UpdateArtworkByMap(ctx, id, map[string]any{"r18": r18})
}

func UpdateArtworkTagsByURL(ctx context.Context, sourceURL string, tags []string) error {
	awEnt, err := database.Default().GetArtworkByURL(ctx, sourceURL)
	if err != nil {
		return err
	}
	uniTags := slice.Unique(tags)
	database.Default().Transaction(ctx, func(tx *database.DB, _ *gorm.DB) error {
		tagEnts := make([]*entity.Tag, 0, len(uniTags))
		for _, tag := range uniTags {
			tagEnt, err := GetTagByNameWithAlias(ctx, tag)
			if err != nil && !errors.Is(err, database.ErrRecordNotFound) {
				return err
			}
			if tagEnt != nil {
				tagEnts = append(tagEnts, tagEnt)
				continue
			}
			tagEnt = &entity.Tag{
				Name: tag,
			}
			res, err := database.Default().CreateTag(ctx, tagEnt)
			if err != nil {
				return err
			}
			tagEnts = append(tagEnts, res)
			// [TODO] tag 排序
		}
		return database.Default().UpdateArtworkTags(ctx, awEnt.ID, tagEnts)
	})
	return err
}
