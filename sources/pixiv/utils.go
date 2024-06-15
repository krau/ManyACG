package pixiv

import (
	"ManyACG/config"
	. "ManyACG/logger"
	"ManyACG/types"
	"encoding/json"
	"encoding/xml"
	"strings"
	"time"
)

func GetPid(url string) string {
	matchUrl := pixivSourceURLRegexp.FindString(url)
	return numberRegexp.FindString(matchUrl)
}

func reqAjaxResp(sourceURL string) (*PixivAjaxResp, error) {
	ajaxURL := "https://www.pixiv.net/ajax/illust/" + GetPid(sourceURL)
	Logger.Tracef("request artwork info: %s", ajaxURL)
	resp, err := ReqClient.R().Get(ajaxURL)
	if err != nil {
		return nil, err
	}
	var pixivAjaxResp PixivAjaxResp
	err = json.Unmarshal(resp.Bytes(), &pixivAjaxResp)
	if err != nil {
		Logger.Errorf("Error decoding artwork info: %v", err)
		return nil, ErrUnmarshalPixivAjax
	}
	return &pixivAjaxResp, nil
}

func reqIllustPages(sourceURL string) (*PixivIllustPages, error) {
	ajaxURL := "https://www.pixiv.net/ajax/illust/" + GetPid(sourceURL) + "/pages?lang=zh"
	Logger.Tracef("request artwork pages: %s", ajaxURL)
	resp, err := ReqClient.R().Get(ajaxURL)
	if err != nil {
		return nil, err
	}
	var pixivIllustPages PixivIllustPages
	err = json.Unmarshal(resp.Bytes(), &pixivIllustPages)
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
		if config.Cfg.Source.Pixiv.Sleep > 0 {
			time.Sleep(time.Duration(config.Cfg.Source.Pixiv.Sleep) * time.Second)
		}
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
		if config.Cfg.Source.Pixiv.Sleep > 0 {
			time.Sleep(time.Duration(config.Cfg.Source.Pixiv.Sleep) * time.Second)
		}
	}
	return artworks, nil
}
