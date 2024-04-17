package pixiv

import (
	"ManyACG-Bot/common"
	. "ManyACG-Bot/logger"
	"ManyACG-Bot/types"
	"context"
	"encoding/json"
	"encoding/xml"
	"strings"
	"sync"

	"golang.org/x/sync/semaphore"
)

func getPid(url string) string {
	return strings.Split(url, "/")[len(strings.Split(url, "/"))-1]
}

func reqAjaxResp(sourceURL string) (*PixivAjaxResp, error) {
	ajaxURL := "https://www.pixiv.net/ajax/illust/" + getPid(sourceURL)
	Logger.Debugf("request artwork info: %s", ajaxURL)
	resp, err := common.Cilent.R().Get(ajaxURL)
	if err != nil {
		return nil, err
	}
	var pixivAjaxResp PixivAjaxResp
	err = json.Unmarshal([]byte(resp.String()), &pixivAjaxResp)
	if err != nil {
		return nil, err
	}
	return &pixivAjaxResp, nil
}

func reqIllustPages(sourceURL string) (*PixivIllustPages, error) {
	ajaxURL := "https://www.pixiv.net/ajax/illust/" + getPid(sourceURL) + "/pages?lang=zh"
	Logger.Debugf("request artwork pages: %s", ajaxURL)
	resp, err := common.Cilent.R().Get(ajaxURL)
	if err != nil {
		return nil, err
	}
	var pixivIllustPages PixivIllustPages
	err = json.Unmarshal([]byte(resp.String()), &pixivIllustPages)
	if err != nil {
		return nil, err
	}
	return &pixivIllustPages, nil
}

func fetchNewArtworksForRSSURL(rssURL string, limit int, wg *sync.WaitGroup, artworkChan chan *types.Artwork) {
	defer wg.Done()
	Logger.Infof("Fetching %s", rssURL)
	resp, err := common.Cilent.R().Get(rssURL)

	if err != nil {
		Logger.Errorf("Error fetching %s: %v", rssURL, err)
		return
	}

	var pixivRss *PixivRss
	err = xml.NewDecoder(strings.NewReader(resp.String())).Decode(&pixivRss)

	if err != nil {
		Logger.Errorf("Error decoding %s: %v", rssURL, err)
		return
	}

	Logger.Debugf("Got %d items", len(pixivRss.Channel.Items))

	// Create a semaphore with a maximum number of concurrent requests
	sem := semaphore.NewWeighted(10) // Set the maximum number of concurrent requests to 10

	for i, item := range pixivRss.Channel.Items {
		if i >= limit {
			break
		}
		wg.Add(1)
		go func(item Item) {
			defer wg.Done()

			// Acquire a semaphore token before making the request
			err := sem.Acquire(context.TODO(), 1)
			if err != nil {
				Logger.Errorf("Error acquiring semaphore token: %v", err)
				return
			}

			// Release the semaphore token after the request is done
			defer sem.Release(1)

			ajaxResp, err := reqAjaxResp(item.Link)
			if err != nil {
				Logger.Errorf("Error fetching artwork info: %v", err)
				return
			}
			artwork, err := ajaxResp.ToArtwork()
			if err != nil {
				Logger.Errorf("Error converting item to artwork: %v", err)
				return
			}
			artworkChan <- artwork
		}(item)
	}
}
