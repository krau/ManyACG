package service

import (
	"context"

	"github.com/krau/ManyACG/internal/common/httpclient"
	"github.com/krau/ManyACG/pkg/osutil"
)

// DownloadOriginFile returns the origin file for url, delegating to the
// configured downloader when set.
func (s *Service) DownloadOriginFile(ctx context.Context, url string) (*osutil.File, error) {
	return httpclient.DownloadWithCacheUsing(ctx, s.downloader, url, nil)
}
