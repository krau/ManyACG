package service

import (
	"context"

	"github.com/krau/ManyACG/internal/common/httpclient"
	"github.com/krau/ManyACG/pkg/osutil"
)

func (s *Service) DownloadOriginFile(ctx context.Context, url string) (*osutil.File, error) {
	return httpclient.DownloadWithCache(ctx, url, s.downloader)
}
