package pixiv

import (
	"context"

	"github.com/goccy/go-json"
	"github.com/imroc/req/v3"
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
