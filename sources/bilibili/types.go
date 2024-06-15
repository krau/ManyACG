package bilibili

import (
	"ManyACG/common"
	"ManyACG/types"
	"errors"
	"fmt"
)

type BilibiliApiResp struct {
	Code    int              `json:"code"`
	Message string           `json:"message"`
	TTL     int              `json:"ttl"`
	Data    *BilibiliApiData `json:"data"`
}

type BilibiliApiData struct {
	Item *BilibiliApiItem `json:"item"`
}

type BilibiliApiItem struct {
	Modules *struct {
		ModuleAuthor  *BilibiliApiModuleAuthor  `json:"module_author"`
		ModuleDynamic *BilibiliApiModuleDynamic `json:"module_dynamic"`
	} `json:"modules"`
	Type  string `json:"type"`
	IdStr string `json:"id_str"`
}

type BilibiliApiModuleAuthor struct {
	Name string `json:"name"`
	Mid  int    `json:"mid"`
}

type BilibiliApiModuleDynamic struct {
	Major *struct {
		Opus *struct {
			Pics    []*BilibiliApiPic   `json:"pics"`
			Summary *BilibiliApiSummary `json:"summary"`
			Title   string              `json:"title"`
		} `json:"opus"`
		Type string `json:"type"`
	} `json:"major"`
}

type BilibiliApiPic struct {
	Height int     `json:"height"`
	Width  int     `json:"width"`
	Size   float64 `json:"size"`
	Url    string  `json:"url"`
}

type BilibiliApiSummary struct {
	Text string `json:"text"`
}

func (resp *BilibiliApiResp) ToArtwork() (*types.Artwork, error) {
	if resp.Code != 0 {
		return nil, errors.New(resp.Message + " (code: " + fmt.Sprint(resp.Code) + ")")
	}
	if resp.Data == nil {
		return nil, ErrInvalidURL
	}
	data := resp.Data
	if data.Item == nil {
		return nil, ErrInvalidURL
	}
	item := data.Item
	if item.Modules.ModuleAuthor == nil || item.Modules.ModuleDynamic == nil || item.Type != "DYNAMIC_TYPE_DRAW" {
		return nil, ErrInvalidURL
	}
	dynamic := item.Modules.ModuleDynamic
	author := item.Modules.ModuleAuthor
	if dynamic.Major == nil || dynamic.Major.Opus == nil {
		return nil, ErrInvalidURL
	}
	opus := dynamic.Major.Opus
	if opus.Pics == nil || opus.Summary == nil {
		return nil, ErrInvalidURL
	}

	pics := opus.Pics
	summary := opus.Summary
	pictures := make([]*types.Picture, 0, len(pics))
	for i, pic := range pics {
		pictures = append(pictures, &types.Picture{
			Index:     uint(i),
			Original:  pic.Url,
			Width:     uint(pic.Width),
			Height:    uint(pic.Height),
			Thumbnail: pic.Url,
		})
	}
	title := opus.Title
	if title == "" {
		title = "bilibili/" + item.IdStr
	}
	artwork := &types.Artwork{
		Title:       title,
		Description: summary.Text,
		SourceType:  types.SourceTypeBilibili,
		SourceURL:   "https://t.bilibili.com/" + item.IdStr,
		R18:         false,
		Artist: &types.Artist{
			Name:     author.Name,
			Username: author.Name,
			Type:     types.SourceTypeBilibili,
			UID:      author.Mid,
		},
		Pictures: pictures,
		Tags:     nil,
	}
	common.DownloadWithCache(artwork.Pictures[0].Original, ReqClient)
	return artwork, nil
}
