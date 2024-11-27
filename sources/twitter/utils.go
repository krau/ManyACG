package twitter

import (
	"context"
	"encoding/json"
	"regexp"
	"strings"
	"time"

	"github.com/imroc/req/v3"
	"github.com/krau/ManyACG/common"
	"github.com/krau/ManyACG/config"

	"github.com/krau/ManyACG/types"
	"github.com/mmcdole/gofeed"
)

var (
	twitterSourceURLRegexp *regexp.Regexp = regexp.MustCompile(`(?:twitter|x)\.com/([^/]+)/status/(\d+)`)
	reqClient              *req.Client
)

func reqApiResp(url string) (*FxTwitterApiResp, error) {
	common.Logger.Tracef("request artwork info: %s", url)
	resp, err := reqClient.R().Get(url)
	if err != nil {
		common.Logger.Errorf("request failed: %v", err)
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

func (t *Twitter) fetchRssURL(url string, limit int) ([]*types.Artwork, error) {
	common.Logger.Infof("Fetching %s", url)
	resp, err := reqClient.R().Get(url)
	if err != nil {
		common.Logger.Errorf("Error fetching %s: %v", url, err)
		return nil, err
	}
	feed, err := gofeed.NewParser().Parse(resp.Body)
	if err != nil {
		common.Logger.Errorf("Error parsing feed: %v", err)
		return nil, err
	}
	common.Logger.Debugf("Got %d items", len(feed.Items))
	artworks := make([]*types.Artwork, 0)
	for i, item := range feed.Items {
		if i >= limit {
			break
		}
		sourceURL := item.Link
		artworkInDB, _ := service.GetArtworkByURL(context.TODO(), sourceURL)
		if artworkInDB != nil {
			common.Logger.Infof("Artwork %s already exists", sourceURL)
			continue
		}
		artwork, err := t.GetArtworkInfo(sourceURL)
		if err != nil {
			common.Logger.Errorf("Error getting artwork info: %v", err)
			continue
		}
		artworks = append(artworks, artwork)
		if config.Cfg.Source.Twitter.Sleep > 0 {
			time.Sleep(time.Duration(config.Cfg.Source.Twitter.Sleep) * time.Second)
		}
	}
	return artworks, nil
}

func (t *Twitter) fetchRssURLWithCh(url string, limit int, artworkCh chan *types.Artwork) error {
	common.Logger.Infof("Fetching %s", url)
	resp, err := reqClient.R().Get(url)
	if err != nil {
		common.Logger.Errorf("Error fetching %s: %v", url, err)
		return err
	}
	feed, err := gofeed.NewParser().Parse(resp.Body)
	if err != nil {
		common.Logger.Errorf("Error parsing feed: %v", err)
		return err
	}
	common.Logger.Debugf("Got %d items", len(feed.Items))
	for i, item := range feed.Items {
		if i >= limit {
			break
		}
		sourceURL := item.Link
		artworkInDB, _ := service.GetArtworkByURL(context.TODO(), sourceURL)
		if artworkInDB != nil {
			common.Logger.Infof("Artwork %s already exists", sourceURL)
			continue
		}
		artwork, err := t.GetArtworkInfo(sourceURL)
		if err != nil {
			common.Logger.Errorf("Error getting artwork info: %v", err)
			continue
		}
		artworkCh <- artwork
		if config.Cfg.Source.Twitter.Sleep > 0 {
			time.Sleep(time.Duration(config.Cfg.Source.Twitter.Sleep) * time.Second)
		}
	}
	return nil
}
