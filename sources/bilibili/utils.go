package bilibili

import (
	"encoding/json"
	"fmt"

	. "github.com/krau/ManyACG/logger"
)

func getDynamicID(url string) string {
	return numberRegexp.FindString(dynamicURLRegexp.FindString(url))
}

func reqApiResp(url string) (*BilibiliApiResp, error) {
	Logger.Tracef("request artwork info: %s", url)
	apiUrl := fmt.Sprintf(apiURLFormat, getDynamicID(url))
	resp, err := ReqClient.R().Get(apiUrl)
	if err != nil {
		Logger.Errorf("request failed: %v", err)
		return nil, ErrRequestFailed
	}
	var bilibiliApiResp BilibiliApiResp
	err = json.Unmarshal(resp.Bytes(), &bilibiliApiResp)
	if err != nil {
		return nil, err
	}
	return &bilibiliApiResp, nil
}
