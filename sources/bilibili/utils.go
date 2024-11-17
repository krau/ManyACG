package bilibili

import (
	"encoding/json"
	"fmt"

	"github.com/krau/ManyACG/common"
)

func getDynamicID(url string) string {
	return numberRegexp.FindString(dynamicURLRegexp.FindString(url))
}

func reqApiResp(url string) (*BilibiliApiResp, error) {
	common.Logger.Tracef("request artwork info: %s", url)
	apiUrl := fmt.Sprintf(apiURLFormat, getDynamicID(url))
	resp, err := reqClient.R().Get(apiUrl)
	if err != nil {
		common.Logger.Errorf("request failed: %v", err)
		return nil, ErrRequestFailed
	}
	var bilibiliApiResp BilibiliApiResp
	err = json.Unmarshal(resp.Bytes(), &bilibiliApiResp)
	if err != nil {
		return nil, err
	}
	return &bilibiliApiResp, nil
}
