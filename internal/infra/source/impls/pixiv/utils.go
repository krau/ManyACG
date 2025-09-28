package pixiv

import (
	"encoding/json"
)

func GetPid(url string) string {
	matchUrl := pixivSourceURLRegexp.FindString(url)
	return numberRegexp.FindString(matchUrl)
}

func reqAjaxResp(sourceURL string) (*PixivAjaxResp, error) {
	ajaxURL := "https://www.pixiv.net/ajax/illust/" + GetPid(sourceURL)
	resp, err := reqClient.R().Get(ajaxURL)
	if err != nil {
		return nil, err
	}
	var pixivAjaxResp PixivAjaxResp
	err = json.Unmarshal(resp.Bytes(), &pixivAjaxResp)
	if err != nil {
		return nil, ErrUnmarshalPixivAjax
	}
	return &pixivAjaxResp, nil
}

func reqIllustPages(sourceURL string) (*PixivIllustPages, error) {
	ajaxURL := "https://www.pixiv.net/ajax/illust/" + GetPid(sourceURL) + "/pages?lang=zh"
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

// func fetchNewArtworksForRSSURLWithCh(rssURL string, limit int, artworkCh chan *types.Artwork) error {
// 	resp, err := reqClient.R().Get(rssURL)

// 	if err != nil {
// 		return err
// 	}

// 	var pixivRss *PixivRss
// 	err = xml.NewDecoder(strings.NewReader(resp.String())).Decode(&pixivRss)

// 	if err != nil {
// 		return err
// 	}


// 	for i, item := range pixivRss.Channel.Items {
// 		if i >= limit {
// 			break
// 		}
// 		artworkInDB, _ := service.GetArtworkByURL(context.TODO(), item.Link)
// 		if artworkInDB != nil {
// 			continue
// 		}
// 		ajaxResp, err := reqAjaxResp(item.Link)
// 		if err != nil {
// 			continue
// 		}
// 		artwork, err := ajaxResp.ToArtwork()
// 		if err != nil {
// 			continue
// 		}
// 		artworkCh <- artwork
// 		if config.Cfg.Source.Pixiv.Sleep > 0 {
// 			time.Sleep(time.Duration(config.Cfg.Source.Pixiv.Sleep) * time.Second)
// 		}
// 	}
// 	return nil
// }

// func fetchNewArtworksForRSSURL(rssURL string, limit int) ([]*types.Artwork, error) {
// 	resp, err := reqClient.R().Get(rssURL)
// 	if err != nil {
// 		return nil, err
// 	}

// 	var pixivRss *PixivRss
// 	err = xml.NewDecoder(strings.NewReader(resp.String())).Decode(&pixivRss)
// 	if err != nil {
// 		return nil, err
// 	}

// 	artworks := make([]*types.Artwork, 0)
// 	for i, item := range pixivRss.Channel.Items {
// 		if i >= limit {
// 			break
// 		}
// 		artworkInDB, _ := service.GetArtworkByURL(context.TODO(), item.Link)
// 		if artworkInDB != nil {
// 			continue
// 		}
// 		ajaxResp, err := reqAjaxResp(item.Link)
// 		if err != nil {
// 			continue
// 		}
// 		artwork, err := ajaxResp.ToArtwork()
// 		if err != nil {
// 			continue
// 		}
// 		artworks = append(artworks, artwork)
// 		if config.Cfg.Source.Pixiv.Sleep > 0 {
// 			time.Sleep(time.Duration(config.Cfg.Source.Pixiv.Sleep) * time.Second)
// 		}
// 	}
// 	return artworks, nil
// }
