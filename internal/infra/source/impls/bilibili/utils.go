package bilibili

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/krau/ManyACG/pkg/reutil"
)

func getDynamicID(url string) string {
	id, ok := reutil.GetLatestNumberFromString(dynamicURLRegexp.FindString(url))
	if !ok {
		return ""
	}
	return id
}

func (b *Bilibili) reqWebDynamicApiResp(ctx context.Context, dynamicID string) (*BilibiliWebDynamicApiResp, error) {
	apiUrl := fmt.Sprintf(webDynamicAPIURLFormat, dynamicID)
	resp, err := b.reqClient.R().SetContext(ctx).Get(apiUrl)
	if err != nil {
		return nil, ErrRequestFailed
	}
	var bilibiliWebDynamicApiResp BilibiliWebDynamicApiResp
	err = json.Unmarshal(resp.Bytes(), &bilibiliWebDynamicApiResp)
	if err != nil {
		return nil, err
	}
	return &bilibiliWebDynamicApiResp, nil
}

func (b *Bilibili) reqDesktopDynamicApiResp(ctx context.Context, dynamicID string) (*BilibiliDesktopDynamicApiResp, error) {
	apiUrl := fmt.Sprintf(desktopDynamicAPIURLFormat, dynamicID)
	resp, err := b.reqClient.R().SetContext(ctx).Get(apiUrl)
	if err != nil {
		return nil, ErrRequestFailed
	}
	var bilibiliDesktopDynamicApiResp BilibiliDesktopDynamicApiResp
	err = json.Unmarshal(resp.Bytes(), &bilibiliDesktopDynamicApiResp)
	if err != nil {
		return nil, err
	}
	return &bilibiliDesktopDynamicApiResp, nil
}
