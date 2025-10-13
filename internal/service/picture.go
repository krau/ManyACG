package service

import (
	"context"

	"github.com/duke-git/lancet/v2/slice"
	"github.com/krau/ManyACG/internal/model/entity"
	"github.com/krau/ManyACG/internal/model/query"
	"github.com/krau/ManyACG/internal/repo"
	"github.com/krau/ManyACG/internal/shared"
	"github.com/krau/ManyACG/pkg/objectuuid"
	"github.com/samber/oops"
	"gorm.io/datatypes"
)

func (s *Service) UpdatePictureTelegramInfo(ctx context.Context, id objectuuid.ObjectUUID, tgInfo *shared.TelegramInfo) error {
	_, err := s.repos.Picture().UpdatePictureTelegramInfoByID(ctx, id, tgInfo)
	return err
}

func (s *Service) QueryPicturesByPhash(ctx context.Context, que query.PicturesPhash) ([]*entity.Picture, error) {
	return s.repos.Picture().QueryPicturesByPhash(ctx, que)
}

// 删除单张图片, 如果删除后对应的 artwork 中没有图片, 则也删除 artwork
//
// 删除后对 artwork 的 pictures 的 index 进行重整, 并在 cached_artwork 中将对应的图片标记为隐藏
func (s *Service) DeletePictureByID(ctx context.Context, id objectuuid.ObjectUUID) error {
	toDelete, err := s.repos.Picture().GetPictureByID(ctx, id)
	if err != nil {
		return err
	}
	artwork, err := s.repos.Artwork().GetArtworkByID(ctx, toDelete.ArtworkID)
	if err != nil {
		return err
	}
	err = s.repos.Transaction(ctx, func(repos repo.Repositories) error {
		if len(artwork.Pictures) == 1 {
			return repos.Artwork().DeleteArtworkByID(ctx, artwork.ID)
		}
		if err := repos.Picture().DeletePictureByID(ctx, id); err != nil {
			return oops.Wrapf(err, "failed to delete picture by id %s", id.String())
		}
		newPictures := slice.Filter(artwork.Pictures, func(index int, item *entity.Picture) bool {
			return item.ID != toDelete.ID
		})
		if err := repos.Artwork().UpdateArtworkPictures(ctx, artwork.ID, newPictures); err != nil {
			return oops.Wrapf(err, "failed to update artwork pictures")
		}
		if err := repos.Artwork().ReorderArtworkPicturesByID(ctx, artwork.ID); err != nil {
			return oops.Wrapf(err, "failed to reorder artwork pictures")
		}
		cached, err := repos.CachedArtwork().GetCachedArtworkByURL(ctx, artwork.SourceURL)
		if err == nil {
			data := cached.Artwork.Data()
			for _, pic := range data.Pictures {
				// 将对应的图片标记为隐藏
				if pic.Original == toDelete.Original {
					pic.Hidden = true
				}
			}
			cached.Artwork = datatypes.NewJSONType(data)
			if _, err := repos.CachedArtwork().SaveCachedArtwork(ctx, cached); err != nil {
				return oops.Wrapf(err, "failed to save cached artwork")
			}
		}
		return nil
	})
	return err
}

func (s *Service) GetPictureByID(ctx context.Context, id objectuuid.ObjectUUID) (*entity.Picture, error) {
	return s.repos.Picture().GetPictureByID(ctx, id)
}

func (s *Service) UpdateUgoiraTelegramInfo(ctx context.Context, id objectuuid.ObjectUUID, tgInfo *shared.TelegramInfo) error {
	_, err := s.repos.Ugoira().UpdateUgoiraTelegramInfoByID(ctx, id, tgInfo)
	return err
}

func (s *Service) RandomPictures(ctx context.Context, limit int) ([]*entity.Picture, error) {
	return s.repos.Picture().RandomPictures(ctx, limit)
}
