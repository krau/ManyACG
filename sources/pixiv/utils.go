package pixiv

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"strings"
	"time"

	"github.com/krau/ManyACG/common"
	"github.com/krau/ManyACG/config"

	"github.com/krau/ManyACG/types"
)

func GetPid(url string) string {
	matchUrl := pixivSourceURLRegexp.FindString(url)
	return numberRegexp.FindString(matchUrl)
}

func reqAjaxResp(sourceURL string) (*PixivAjaxResp, error) {
	ajaxURL := "https://www.pixiv.net/ajax/illust/" + GetPid(sourceURL)
	common.Logger.Tracef("request artwork info: %s", ajaxURL)
	resp, err := reqClient.R().Get(ajaxURL)
	if err != nil {
		return nil, err
	}
	var pixivAjaxResp PixivAjaxResp
	err = json.Unmarshal(resp.Bytes(), &pixivAjaxResp)
	if err != nil {
		common.Logger.Errorf("Error decoding artwork info: %v", err)
		return nil, ErrUnmarshalPixivAjax
	}
	return &pixivAjaxResp, nil
}

func reqIllustPages(sourceURL string) (*PixivIllustPages, error) {
	ajaxURL := "https://www.pixiv.net/ajax/illust/" + GetPid(sourceURL) + "/pages?lang=zh"
	common.Logger.Tracef("request artwork pages: %s", ajaxURL)
	resp, err := reqClient.R().Get(ajaxURL)
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
	common.Logger.Infof("Fetching %s", rssURL)
	resp, err := reqClient.R().Get(rssURL)

	if err != nil {
		common.Logger.Errorf("Error fetching %s: %v", rssURL, err)
		return err
	}

	var pixivRss *PixivRss
	err = xml.NewDecoder(strings.NewReader(resp.String())).Decode(&pixivRss)

	if err != nil {
		common.Logger.Errorf("Error decoding %s: %v", rssURL, err)
		return err
	}

	common.Logger.Infof("Got %d items", len(pixivRss.Channel.Items))

	for i, item := range pixivRss.Channel.Items {
		if i >= limit {
			break
		}
		artworkInDB, _ := service.GetArtworkByURL(context.TODO(), item.Link)
		if artworkInDB != nil {
			common.Logger.Infof("Artwork %s already exists", item.Link)
			continue
		}
		ajaxResp, err := reqAjaxResp(item.Link)
		if err != nil {
			common.Logger.Errorf("Error fetching artwork info: %v", err)
			continue
		}
		artwork, err := ajaxResp.ToArtwork()
		if err != nil {
			common.Logger.Errorf("Error converting item to artwork: %v", err)
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
	common.Logger.Infof("Fetching %s", rssURL)
	resp, err := reqClient.R().Get(rssURL)
	if err != nil {
		common.Logger.Errorf("Error fetching %s: %v", rssURL, err)
		return nil, err
	}

	var pixivRss *PixivRss
	err = xml.NewDecoder(strings.NewReader(resp.String())).Decode(&pixivRss)
	if err != nil {
		common.Logger.Errorf("Error decoding %s: %v", rssURL, err)
		return nil, err
	}

	common.Logger.Debugf("Got %d items", len(pixivRss.Channel.Items))
	artworks := make([]*types.Artwork, 0)
	for i, item := range pixivRss.Channel.Items {
		if i >= limit {
			break
		}
		artworkInDB, _ := service.GetArtworkByURL(context.TODO(), item.Link)
		if artworkInDB != nil {
			common.Logger.Infof("Artwork %s already exists", item.Link)
			continue
		}
		ajaxResp, err := reqAjaxResp(item.Link)
		if err != nil {
			common.Logger.Errorf("Error fetching artwork info: %v", err)
			continue
		}
		artwork, err := ajaxResp.ToArtwork()
		if err != nil {
			common.Logger.Errorf("Error converting item to artwork: %v", err)
			continue
		}
		artworks = append(artworks, artwork)
		if config.Cfg.Source.Pixiv.Sleep > 0 {
			time.Sleep(time.Duration(config.Cfg.Source.Pixiv.Sleep) * time.Second)
		}
	}
	return artworks, nil
}
