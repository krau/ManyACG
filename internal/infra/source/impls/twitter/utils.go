package twitter

import (
	"context"
	"encoding/json"
	"regexp"

	"github.com/imroc/req/v3"
)

var (
	twitterSourceURLRegexp *regexp.Regexp = regexp.MustCompile(`(?:twitter|x)\.com/([^/]+)/status/(\d+)`)
	reqClient              *req.Client
)

func reqApiResp(ctx context.Context, url string) (*FxTwitterApiResp, error) {
	resp, err := reqClient.R().SetContext(ctx).Get(url)
	if err != nil {
		return nil, ErrRequestFailed
	}
	var fxTwitterApiResp FxTwitterApiResp
	err = json.Unmarshal(resp.Bytes(), &fxTwitterApiResp)
	if err != nil {
		return nil, err
	}
	return &fxTwitterApiResp, nil
}

func getTweetID(sourceURL string) string {
	matches := twitterSourceURLRegexp.FindStringSubmatch(sourceURL)
	if len(matches) < 3 {
		return ""
	}
	return matches[2]
}

func getTweetPath(sourceURL string) string {
	matches := twitterSourceURLRegexp.FindStringSubmatch(sourceURL)
	if len(matches) < 3 {
		return ""
	}
	return matches[1] + "/status/" + matches[2]
}

// func (t *Twitter) fetchRssURL(url string, limit int) ([]*types.Artwork, error) {
// 	resp, err := reqClient.R().Get(url)
// 	if err != nil {
// 		return nil, err
// 	}
// 	feed, err := gofeed.NewParser().Parse(resp.Body)
// 	if err != nil {
// 		return nil, err
// 	}
// 	artworks := make([]*types.Artwork, 0)
// 	for i, item := range feed.Items {
// 		if i >= limit {
// 			break
// 		}
// 		sourceURL := item.Link
// 		artworkInDB, _ := service.GetArtworkByURL(context.TODO(), sourceURL)
// 		if artworkInDB != nil {
// 			continue
// 		}
// 		artwork, err := t.GetArtworkInfo(sourceURL)
// 		if err != nil {
// 			continue
// 		}
// 		artworks = append(artworks, artwork)
// 		if config.Get().Source.Twitter.Sleep > 0 {
// 			time.Sleep(time.Duration(config.Get().Source.Twitter.Sleep) * time.Second)
// 		}
// 	}
// 	return artworks, nil
// }

// func (t *Twitter) fetchRssURLWithCh(url string, limit int, artworkCh chan *types.Artwork) error {
// 	resp, err := reqClient.R().Get(url)
// 	if err != nil {
// 		return err
// 	}
// 	feed, err := gofeed.NewParser().Parse(resp.Body)
// 	if err != nil {
// 		return err
// 	}
// 	for i, item := range feed.Items {
// 		if i >= limit {
// 			break
// 		}
// 		sourceURL := item.Link
// 		artworkInDB, _ := service.GetArtworkByURL(context.TODO(), sourceURL)
// 		if artworkInDB != nil {
// 			continue
// 		}
// 		artwork, err := t.GetArtworkInfo(sourceURL)
// 		if err != nil {
// 			continue
// 		}
// 		artworkCh <- artwork
// 		if config.Get().Source.Twitter.Sleep > 0 {
// 			time.Sleep(time.Duration(config.Get().Source.Twitter.Sleep) * time.Second)
// 		}
// 	}
// 	return nil
// }
