package bilibili

import (
	"encoding/json"
	"fmt"
)

func getDynamicID(url string) string {
	return numberRegexp.FindString(dynamicURLRegexp.FindString(url))
}

func reqWebDynamicApiResp(dynamicID string) (*BilibiliWebDynamicApiResp, error) {
	apiUrl := fmt.Sprintf(webDynamicAPIURLFormat, dynamicID)
	resp, err := reqClient.R().Get(apiUrl)
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

func reqDesktopDynamicApiResp(dynamicID string) (*BilibiliDesktopDynamicApiResp, error) {
	apiUrl := fmt.Sprintf(desktopDynamicAPIURLFormat, dynamicID)
	resp, err := reqClient.R().Get(apiUrl)
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
