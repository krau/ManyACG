package service

import (
	"context"

	"github.com/krau/ManyACG/dao"
	"github.com/krau/ManyACG/types"
)

func DeleteDeletedByURL(ctx context.Context, sourceURL string) error {
	_, err := dao.DeleteDeletedByURL(ctx, sourceURL)
	if err != nil {
		return err
	}
	return nil
}

func CheckDeletedByURL(ctx context.Context, sourceURL string) bool {
	return dao.CheckDeletedByURL(ctx, sourceURL)
}

func GetDeletedByURL(ctx context.Context, sourceURL string) (*types.DeletedModel, error) {
	return dao.GetDeletedByURL(ctx, sourceURL)
}
