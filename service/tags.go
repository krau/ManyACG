package service

import (
	"ManyACG/dao"
	"ManyACG/model"
	"context"
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

func GetRandomTagModels(ctx context.Context, limit int) ([]*model.TagModel, error) {
	tags, err := dao.GetRandomTags(ctx, limit)
	if err != nil {
		return nil, err
	}
	return tags, nil
}
