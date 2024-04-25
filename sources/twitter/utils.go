package twitter

import (
	"ManyACG-Bot/common"
	"encoding/json"
	"strings"
)

func reqApiResp(url string) (*FxTwitterApiResp, error) {
	resp, err := common.Client.R().Get(url)
	if err != nil {
		return nil, err
	}
	var fxTwitterApiResp FxTwitterApiResp
	err = json.Unmarshal(resp.Bytes(), &fxTwitterApiResp)
	if err != nil {
		return nil, err
	}
	return &fxTwitterApiResp, nil
}

func GetTweetPath(sourceURL string) (string, error) {
	urlSplit := strings.Split(sourceURL, "/")
	if len(urlSplit) < 6 {
		return "", ErrInvalidURL
	}
	authorUsername := urlSplit[3]
	tweetID := urlSplit[5]
	tweetID = strings.Split(tweetID, "?")[0]
	tweetPath := authorUsername + "/status/" + tweetID
	return tweetPath, nil
}
