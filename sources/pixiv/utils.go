package pixiv

import (
	"ManyACG-Bot/common"
	"ManyACG-Bot/model"
	"encoding/json"
	"encoding/xml"
	"strings"
	"sync"

	. "ManyACG-Bot/logger"
)

func getArtworkInfo(sourceURL string) (*PixivAjaxResp, error) {
	pid := strings.Split(sourceURL, "/")[len(strings.Split(sourceURL, "/"))-1]
	ajaxURL := "https://www.pixiv.net/ajax/illust/" + pid
	Logger.Debugf("Fetching artwork info: %s", ajaxURL)
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

func getNewArtworksForURL(url string, limit int, wg *sync.WaitGroup, artworkChan chan *model.Artwork) {
	defer wg.Done()
	Logger.Infof("Fetching %s", url)
	resp, err := common.Cilent.R().Get(url)

	if err != nil {
		Logger.Errorf("Error fetching %s: %v", url, err)
		return
	}

	var pixivRss *PixivRss
	err = xml.NewDecoder(strings.NewReader(resp.String())).Decode(&pixivRss)

	if err != nil {
		Logger.Errorf("Error decoding %s: %v", url, err)
		return
	}

	Logger.Debugf("Got %d items", len(pixivRss.Channel.Items))

	for i, item := range pixivRss.Channel.Items {
		if i >= limit {
			break
		}
		wg.Add(1)
		go func(item Item) {
			defer wg.Done()
			artworkInfo, err := getArtworkInfo(item.Link)
			if err != nil {
				Logger.Errorf("Error fetching artwork info: %v", err)
				return
			}

			artworkChan <- item.ToArtwork(artworkInfo)
		}(item)
	}
}
