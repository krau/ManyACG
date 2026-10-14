package pixiv

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/xml"
	"fmt"
	"strings"

	config "github.com/krau/ManyACG/internal/infra/config/runtimecfg"
	"github.com/krau/ManyACG/internal/infra/kvstor"
	"github.com/krau/ManyACG/internal/model/dto"
	"github.com/krau/ManyACG/pkg/log"

	"github.com/imroc/req/v3"
)

// Poller fetches new artworks from the configured Pixiv RSS feeds.
type Poller struct {
	reqClient *req.Client
	cfg       config.SourcePixivConfig
	kv        kvstor.KVStore
}

func (p *Poller) FetchNewArtworks(ctx context.Context, limit int) ([]*dto.FetchedArtwork, error) {
	artworks := make([]*dto.FetchedArtwork, 0)
	errs := make([]error, 0)
	for _, url := range p.cfg.RssURLs {
		artworksForURL, err := p.fetchNewArtworksForRSSURL(ctx, url, limit)
		if err != nil {
			errs = append(errs, err)
		}
		artworks = append(artworks, artworksForURL...)
	}
	if len(errs) > 0 {
		return artworks, fmt.Errorf("fetching pixiv encountered %d errors: %v", len(errs), errs)
	}
	return artworks, nil
}

func (p *Poller) fetchNewArtworksForRSSURL(ctx context.Context, rssURL string, limit int) ([]*dto.FetchedArtwork, error) {
	resp, err := p.reqClient.R().SetContext(ctx).Get(rssURL)
	if err != nil {
		return nil, err
	}

	body := resp.String()
	rsssum := sha256.Sum256([]byte(body))
	fingerprint := hex.EncodeToString(rsssum[:])
	cacheKey := pixivRSSCacheKey(rssURL)

	if cacheEntry, err := kvstor.Get[pixivRSSCacheEntry](ctx, p.kv, cacheKey); err == nil {
		if cacheEntry.Signature == fingerprint {
			if limit > 0 && len(cacheEntry.Artworks) > limit {
				return cacheEntry.Artworks[:limit], nil
			}
			return cacheEntry.Artworks, nil
		}
	}

	var pixivRss *PixivRss
	if err := xml.NewDecoder(strings.NewReader(body)).Decode(&pixivRss); err != nil {
		return nil, err
	}

	artworks := make([]*dto.FetchedArtwork, 0)
	for i, item := range pixivRss.Channel.Items {
		if i >= limit {
			break
		}
		ajaxResp, err := reqAjaxResp(ctx, item.Link, p.reqClient)
		if err != nil {
			continue
		}
		artwork, err := ajaxResp.ToArtwork(ctx, p.reqClient, p.cfg.ImgProxy)
		if err != nil {
			continue
		}
		artworks = append(artworks, artwork)
	}

	if len(artworks) > 0 || len(pixivRss.Channel.Items) == 0 {
		entry := pixivRSSCacheEntry{Signature: fingerprint, Artworks: artworks}
		if err := p.kv.Set(ctx, cacheKey, entry, 0); err != nil {
			log.Warn("pixiv rss cache store failed", "url", rssURL, "err", err)
		}
	}

	return artworks, nil
}

func pixivRSSCacheKey(rssURL string) string {
	sum := sha256.Sum256([]byte(rssURL))
	return "pixiv:rss:" + hex.EncodeToString(sum[:])
}

type pixivRSSCacheEntry struct {
	Signature string
	Artworks  []*dto.FetchedArtwork
}
