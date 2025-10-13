package database

import (
	"context"
	"errors"
	"fmt"

	"github.com/krau/ManyACG/internal/model/entity"
	"github.com/krau/ManyACG/pkg/objectuuid"
	"gorm.io/gorm"
)

func (d *DB) GetTagByName(ctx context.Context, name string) (*entity.Tag, error) {
	tag, err := gorm.G[entity.Tag](d.db).
		Where("name = ?", name).
		Preload("Alias", nil).
		First(ctx)
	if err != nil {
		return nil, err
	}
	return &tag, nil
}

func (d *DB) GetAliasTagByName(ctx context.Context, name string) (*entity.TagAlias, error) {
	alias, err := gorm.G[entity.TagAlias](d.db).
		Where("alias = ?", name).
		First(ctx)
	if err != nil {
		return nil, err
	}
	return &alias, nil
}

func (d *DB) GetTagByID(ctx context.Context, id objectuuid.ObjectUUID) (*entity.Tag, error) {
	tag, err := gorm.G[entity.Tag](d.db).
		Where("id = ?", id).
		Preload("Alias", nil).
		First(ctx)
	if err != nil {
		return nil, err
	}
	return &tag, nil
}

func (d *DB) GetTagByNameWithAlias(ctx context.Context, name string) (*entity.Tag, error) {
	tag, err := gorm.G[entity.Tag](d.db).
		Where("name = ?", name).
		Preload("Alias", nil).
		First(ctx)
	if err == nil {
		return &tag, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	alias, err := gorm.G[entity.TagAlias](d.db).
		Where("alias = ?", name).
		First(ctx)
	if err != nil {
		return nil, err
	}
	tag, err = gorm.G[entity.Tag](d.db).
		Where("id = ?", alias.TagID).
		Preload("Alias", nil).
		First(ctx)
	if err != nil {
		return nil, err
	}
	return &tag, nil
}

func (d *DB) CreateTag(ctx context.Context, tag *entity.Tag) (*entity.Tag, error) {
	result := gorm.WithResult()
	err := gorm.G[entity.Tag](d.db, result).Create(ctx, tag)
	if err != nil {
		return nil, err
	}
	return tag, nil
}

func (d *DB) UpdateTagAlias(ctx context.Context, id objectuuid.ObjectUUID, alias []*entity.TagAlias) error {
	var tag entity.Tag
	err := d.db.WithContext(ctx).Model(&entity.Tag{}).Where("id = ?", id).First(&tag).Error
	if err != nil {
		return err
	}
	return d.db.WithContext(ctx).Model(&tag).Association("Alias").Replace(alias)
}

func (d *DB) DeleteTagByID(ctx context.Context, id objectuuid.ObjectUUID) error {
	n, err := gorm.G[entity.Tag](d.db).Where("id = ?", id).Delete(ctx)
	if err != nil {
		return err
	}
	if n == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// MigrateTagAlias 迁移别名标签到目标标签，并删除别名标签
//
// 返回受影响的 artwork ids
func (d *DB) MigrateTagAlias(ctx context.Context, aliasTagID, targetTagID objectuuid.ObjectUUID) ([]objectuuid.ObjectUUID, error) {
	var affected []objectuuid.ObjectUUID

	err := d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 所有拥有 aliasTagID 的 artwork_id, 即受影响的 artworks
		rows, err := tx.Raw(`SELECT DISTINCT artwork_id FROM artwork_tags WHERE tag_id = ?`, aliasTagID).Rows()
		if err != nil {
			return fmt.Errorf("select affected artworks: %w", err)
		}
		defer rows.Close()
		for rows.Next() {
			var aid objectuuid.ObjectUUID
			if err := rows.Scan(&aid); err != nil {
				return fmt.Errorf("scan artwork id: %w", err)
			}
			affected = append(affected, aid)
		}

		if err := tx.Exec(`
			INSERT INTO artwork_tags (artwork_id, tag_id)
			SELECT artwork_id, ?
			FROM artwork_tags
			WHERE tag_id = ?
			AND artwork_id NOT IN (
				SELECT artwork_id FROM artwork_tags WHERE tag_id = ?
			)
		`, targetTagID, aliasTagID, targetTagID).Error; err != nil {
			return fmt.Errorf("insert new tag references: %w", err)
		}

		if err := tx.Exec(`DELETE FROM artwork_tags WHERE tag_id = ?`, aliasTagID).Error; err != nil {
			return fmt.Errorf("delete old tag references: %w", err)
		}

		res := tx.Where("id = ?", aliasTagID).Delete(&entity.Tag{})
		if res.Error != nil {
			return fmt.Errorf("delete alias tag: %w", res.Error)
		}
		if res.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return affected, nil
}

// func (d *DB) MigrateTagAlias(ctx context.Context, aliasTagID, targetTagID objectuuid.ObjectUUID) error {
// 	return d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
// 		// 迁移 aliasTagID 的作品引用到 targetTagID（去重）
// 		if err := tx.Exec(`
// 			INSERT INTO artwork_tags (artwork_id, tag_id)
// 			SELECT artwork_id, ?
// 			FROM artwork_tags
// 			WHERE tag_id = ?
// 			AND artwork_id NOT IN (
// 				SELECT artwork_id FROM artwork_tags WHERE tag_id = ?
// 			)
// 		`, targetTagID, aliasTagID, targetTagID).Error; err != nil {
// 			return fmt.Errorf("insert new tag references: %w", err)
// 		}

// 		// 删除 aliasTagID 的旧关联
// 		if err := tx.Exec(`DELETE FROM artwork_tags WHERE tag_id = ?`, aliasTagID).Error; err != nil {
// 			return fmt.Errorf("delete old tag references: %w", err)
// 		}

// 		// 删除 tag 表中 aliasTagID 的记录
// 		res := tx.Where("id = ?", aliasTagID).Delete(&entity.Tag{})
// 		if res.Error != nil {
// 			return fmt.Errorf("delete alias tag: %w", res.Error)
// 		}
// 		if res.RowsAffected == 0 {
// 			return gorm.ErrRecordNotFound
// 		}
// 		return nil
// 	})
// }

func (d *DB) RandomTags(ctx context.Context, limit int) ([]*entity.Tag, error) {
	var tags []*entity.Tag
	err := d.db.WithContext(ctx).Model(&entity.Tag{}).
		Preload("Alias", nil).
		Order("RANDOM()").
		Limit(limit).
		Find(&tags).Error
	if err != nil {
		return nil, err
	}
	return tags, nil
}
