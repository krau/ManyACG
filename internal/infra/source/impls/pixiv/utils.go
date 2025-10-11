package pixiv

import (
	"context"
	"encoding/xml"
	"fmt"
	"strings"

	"github.com/goccy/go-json"
	"github.com/imroc/req/v3"
	"github.com/samber/oops"

	"github.com/krau/ManyACG/internal/infra/cache"
	"github.com/krau/ManyACG/internal/model/dto"
	"github.com/krau/ManyACG/pkg/reutil"
)

func getPid(url string) string {
	matchUrl := sourceReg.FindString(url)
	id, ok := reutil.GetLatestNumberFromString(matchUrl)
	if !ok {
		return ""
	}
	return id
}

func cacheKeyForAjaxResp(sourceURL string) string {
	return fmt.Sprintf("pixiv:reqAjaxResp:%s", sourceURL)
}

func cacheKeyForIllustPages(sourceURL string) string {
	return fmt.Sprintf("pixiv:reqIllustPages:%s", sourceURL)
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
	value, err := cache.Get[PixivAjaxResp](ctx, cacheKeyForAjaxResp(sourceURL))
	if err == nil {
		return &value, nil
	}
	resp, err := doReqAjaxResp(ctx, sourceURL, client)
	if err != nil {
		return nil, err
	}
	cache.Set(ctx, cacheKeyForAjaxResp(sourceURL), *resp)
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
	value, err := cache.Get[PixivIllustPages](ctx, cacheKeyForIllustPages(sourceURL))
	if err == nil {
		return &value, nil
	}
	resp, err := doReqIllustPages(ctx, sourceURL, client)
	if err != nil {
		return nil, err
	}
	cache.Set(ctx, cacheKeyForIllustPages(sourceURL), *resp)
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
	value, err := cache.Get[PixivUgoiraMeta](ctx, cacheKeyForIllustPages(sourceURL)+"-ugoira")
	if err == nil {
		return &value, nil
	}
	resp, err := doReqUgoiraMeta(ctx, sourceURL, client)
	if err != nil {
		return nil, err
	}
	cache.Set(ctx, cacheKeyForIllustPages(sourceURL)+"-ugoira", *resp)
	return resp, nil
}

func (p *Pixiv) fetchNewArtworksForRSSURL(ctx context.Context, rssURL string, limit int) ([]*dto.FetchedArtwork, error) {
	resp, err := p.reqClient.R().SetContext(ctx).Get(rssURL)
	if err != nil {
		return nil, err
	}

	var pixivRss *PixivRss
	err = xml.NewDecoder(strings.NewReader(resp.String())).Decode(&pixivRss)
	if err != nil {
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
	return artworks, nil
}
