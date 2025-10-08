package database

import (
	"context"
	"strings"

	"github.com/corona10/goimagehash"
	"github.com/krau/ManyACG/internal/model/entity"
	"github.com/krau/ManyACG/internal/model/query"
	"github.com/krau/ManyACG/internal/shared"
	"github.com/krau/ManyACG/pkg/objectuuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

func (d *DB) GetPictureByID(ctx context.Context, id objectuuid.ObjectUUID) (*entity.Picture, error) {
	pic, err := gorm.G[entity.Picture](d.db).Where("id = ?", id).First(ctx)
	if err != nil {
		return nil, err
	}
	return &pic, nil
}

// 在数据库删除单张图片, 不做任何额外操作
func (d *DB) DeletePictureByID(ctx context.Context, id objectuuid.ObjectUUID) error {
	n, err := gorm.G[entity.Picture](d.db).Where("id = ?", id).Delete(ctx)
	if err != nil {
		return err
	}
	if n == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (d *DB) ReorderArtworkPicturesByID(ctx context.Context, artworkID objectuuid.ObjectUUID) error {
	// 将 artwork 的 pictures 的 index 重设为连续的数字
	var pictures []entity.Picture
	err := d.db.WithContext(ctx).Model(&entity.Picture{}).Where("artwork_id = ?", artworkID).Order("`index` ASC").Find(&pictures).Error
	if err != nil {
		return err
	}
	for i := range pictures {
		pictures[i].Index = uint(i + 1)
	}
	return d.db.WithContext(ctx).Save(&pictures).Error
}

func (d *DB) QueryPicturesByPhash(ctx context.Context, que query.PicturesPhash) ([]*entity.Picture, error) {
	input := que.Input
	inputHash, err := goimagehash.LoadImageHash(strings.NewReader(input))
	if err != nil {
		return nil, err
	}
	rows, err := d.db.WithContext(ctx).Model(&entity.Picture{}).
		Where("phash IS NOT NULL AND phash <> ''").Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*entity.Picture
	for rows.Next() {
		var pic entity.Picture
		if err := d.db.ScanRows(rows, &pic); err != nil {
			return nil, err
		}
		if pic.Phash == "" {
			continue
		}
		picHash, err := goimagehash.LoadImageHash(strings.NewReader(pic.Phash))
		if err != nil {
			continue
		}
		distance, err := inputHash.Distance(picHash)
		if err != nil {
			continue
		}
		if distance <= que.Distance {
			result = append(result, &pic)
			if que.Limit > 0 && len(result) >= que.Limit {
				break
			}
		}
	}
	return result, nil
}

func (d *DB) UpdatePictureTelegramInfoByID(ctx context.Context, id objectuuid.ObjectUUID, tgInfo *shared.TelegramInfo) (*entity.Picture, error) {
	pic, err := d.GetPictureByID(ctx, id)
	if err != nil {
		return nil, err
	}
	pic.TelegramInfo = datatypes.NewJSONType(*tgInfo)
	err = d.db.WithContext(ctx).Save(pic).Error
	if err != nil {
		return nil, err
	}
	return pic, nil
}
