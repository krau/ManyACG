package twitter

import (
	"encoding/json"
	"regexp"
	"strings"

	"github.com/imroc/req/v3"
	. "github.com/krau/ManyACG/logger"
)

var (
	twitterSourceURLRegexp *regexp.Regexp = regexp.MustCompile(`(?:twitter|x)\.com/([^/]+)/status/(\d+)`)
	reqClient              *req.Client
)

func reqApiResp(url string) (*FxTwitterApiResp, error) {
	Logger.Tracef("request artwork info: %s", url)
	resp, err := reqClient.R().Get(url)
	if err != nil {
		Logger.Errorf("request failed: %v", err)
		return nil, ErrRequestFailed
	}
	var fxTwitterApiResp FxTwitterApiResp
	err = json.Unmarshal(resp.Bytes(), &fxTwitterApiResp)
	if err != nil {
		return nil, err
	}
	return &fxTwitterApiResp, nil
}

func GetTweetPath(sourceURL string) string {
	url := twitterSourceURLRegexp.FindString(sourceURL)
	url = strings.TrimPrefix(url, "twitter.com/")
	url = strings.TrimPrefix(url, "x.com/")
	return url
}
