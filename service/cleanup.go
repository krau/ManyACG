package service

import (
	"context"
	"fmt"

	"github.com/krau/ManyACG/dao"
)

func Cleanup(ctx context.Context) error {
	var errs []error
	if _, err := dao.CleanPostingCachedArtwork(ctx); err != nil {
		errs = append(errs, err)
	}
	if len(errs) > 0 {
		return fmt.Errorf("failed to cleanup: %v", errs)
	}
	return nil
}
