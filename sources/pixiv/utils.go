package pixiv

import (
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
	resp, err := ReqClient.R().Get(ajaxURL)
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
	resp, err := ReqClient.R().Get(ajaxURL)
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

func fetchNewArtworksForRSSURLWithCh(rssURL string, limit int, artworkCh chan *types.Artwork) error {
	Logger.Infof("Fetching %s", rssURL)
	resp, err := ReqClient.R().Get(rssURL)

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

	Logger.Infof("Got %d items", len(pixivRss.Channel.Items))

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

func fetchNewArtworksForRSSURL(rssURL string, limit int) ([]*types.Artwork, error) {
	Logger.Infof("Fetching %s", rssURL)
	resp, err := ReqClient.R().Get(rssURL)
	if err != nil {
		Logger.Errorf("Error fetching %s: %v", rssURL, err)
		return nil, err
	}

	var pixivRss *PixivRss
	err = xml.NewDecoder(strings.NewReader(resp.String())).Decode(&pixivRss)
	if err != nil {
		Logger.Errorf("Error decoding %s: %v", rssURL, err)
		return nil, err
	}

	Logger.Debugf("Got %d items", len(pixivRss.Channel.Items))
	artworks := make([]*types.Artwork, 0)
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
		artworks = append(artworks, artwork)
	}
	return artworks, nil
}
