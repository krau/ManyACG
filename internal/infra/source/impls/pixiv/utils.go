package pixiv

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/xml"
	"strings"

	"github.com/goccy/go-json"
	"github.com/imroc/req/v3"
	"github.com/krau/ManyACG/internal/infra/kvstor"
	"github.com/krau/ManyACG/internal/model/dto"
	"github.com/krau/ManyACG/pkg/log"
	"github.com/krau/ManyACG/pkg/reutil"
	"github.com/samber/oops"
)

func getPid(url string) string {
	matchUrl := sourceReg.FindString(url)
	id, ok := reutil.GetLatestNumberFromString(matchUrl)
	if !ok {
		return ""
	}
	return id
}

func doReqAjaxResp(ctx context.Context, sourceURL string, client *req.Client) (*PixivAjaxResp, error) {
	id := getPid(sourceURL)
	if id == "" {
		return nil, oops.New("invalid pixiv URL, cannot find artwork ID")
	}
	ajaxURL := "https://www.pixiv.net/ajax/illust/" + id
	resp, err := client.R().SetContext(ctx).Get(ajaxURL)
	if err != nil {
		return nil, err
	}
	var pixivAjaxResp PixivAjaxResp
	err = json.Unmarshal(resp.Bytes(), &pixivAjaxResp)
	if err != nil {
		return nil, ErrUnmarshalPixivAjax
	}
	return &pixivAjaxResp, nil
}

func reqAjaxResp(ctx context.Context, sourceURL string, client *req.Client) (*PixivAjaxResp, error) {
	resp, err := doReqAjaxResp(ctx, sourceURL, client)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func doReqIllustPages(ctx context.Context, sourceURL string, client *req.Client) (*PixivIllustPages, error) {
	ajaxURL := "https://www.pixiv.net/ajax/illust/" + getPid(sourceURL) + "/pages?lang=zh"
	resp, err := client.R().SetContext(ctx).Get(ajaxURL)
	if err != nil {
		return nil, err
	}
	var pixivIllustPages PixivIllustPages
	err = json.Unmarshal(resp.Bytes(), &pixivIllustPages)
	if err != nil {
		return nil, err
	}
	return &pixivIllustPages, nil
}

func reqIllustPages(ctx context.Context, sourceURL string, client *req.Client) (*PixivIllustPages, error) {
	resp, err := doReqIllustPages(ctx, sourceURL, client)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func doReqUgoiraMeta(ctx context.Context, sourceURL string, client *req.Client) (*PixivUgoiraMeta, error) {
	ajaxURL := "https://www.pixiv.net/ajax/illust/" + getPid(sourceURL) + "/ugoira_meta?lang=zh"
	resp, err := client.R().SetContext(ctx).Get(ajaxURL)
	if err != nil {
		return nil, err
	}
	var pixivUgoiraMeta PixivUgoiraMeta
	err = json.Unmarshal(resp.Bytes(), &pixivUgoiraMeta)
	if err != nil {
		return nil, err
	}
	return &pixivUgoiraMeta, nil
}

func reqUgoiraMeta(ctx context.Context, sourceURL string, client *req.Client) (*PixivUgoiraMeta, error) {
	resp, err := doReqUgoiraMeta(ctx, sourceURL, client)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (p *Pixiv) fetchNewArtworksForRSSURL(ctx context.Context, rssURL string, limit int) ([]*dto.FetchedArtwork, error) {
	resp, err := p.reqClient.R().SetContext(ctx).Get(rssURL)
	if err != nil {
		return nil, err
	}

	body := resp.String()
	rsssum := sha256.Sum256([]byte(body))
	fingerprint := hex.EncodeToString(rsssum[:])
	cacheKey := pixivRSSCacheKey(rssURL)

	if cacheEntry, err := kvstor.Get[pixivRSSCacheEntry](ctx, cacheKey); err == nil {
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
		if err := kvstor.Set(ctx, cacheKey, entry); err != nil {
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
