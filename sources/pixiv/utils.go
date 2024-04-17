package pixiv

import (
	"ManyACG-Bot/common"
	. "ManyACG-Bot/logger"
	"ManyACG-Bot/types"
	"encoding/json"
	"encoding/xml"
	"strings"
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

func fetchNewArtworksForRSSURL(rssURL string, limit int, artworkCh chan *types.Artwork) error {
	Logger.Infof("Fetching %s", rssURL)
	resp, err := common.Cilent.R().Get(rssURL)

	if err != nil {
		Logger.Errorf("Error fetching %s: %v", rssURL, err)
		return err
	}

	var pixivRss *PixivRss
	err = xml.NewDecoder(strings.NewReader(resp.String())).Decode(&pixivRss)

	if err != nil {
		Logger.Errorf("Error decoding %s: %v", rssURL, err)
		return err
	}

	Logger.Debugf("Got %d items", len(pixivRss.Channel.Items))

	for i, item := range pixivRss.Channel.Items {
		if i >= limit {
			break
		}
		ajaxResp, err := reqAjaxResp(item.Link)
		if err != nil {
			Logger.Errorf("Error fetching artwork info: %v", err)
			continue
		}
		artwork, err := ajaxResp.ToArtwork()
		if err != nil {
			Logger.Errorf("Error converting item to artwork: %v", err)
			continue
		}
		artworkCh <- artwork
	}
	return nil
}
