package bilibili

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/krau/ManyACG/internal/model/dto"
	"github.com/krau/ManyACG/internal/shared"
	"github.com/krau/ManyACG/pkg/strutil"
)

type BilibiliWebDynamicApiResp struct {
	Data    *BilibiliWebDynamicApiData `json:"data"`
	Message string                     `json:"message"`
	Code    int                        `json:"code"`
	TTL     int                        `json:"ttl"`
}

type BilibiliWebDynamicApiData struct {
	Item *BilibiliWebDynamicApiItem `json:"item"`
}

type BilibiliWebDynamicApiItem struct {
	Modules *struct {
		ModuleAuthor  *BilibiliWebDynamicApiModuleAuthor  `json:"module_author"`
		ModuleDynamic *BilibiliWebDynamicApiModuleDynamic `json:"module_dynamic"`
	} `json:"modules"`
	Type  string `json:"type"`
	IdStr string `json:"id_str"`
}

type BilibiliWebDynamicApiModuleAuthor struct {
	Name string `json:"name"`
	Mid  int    `json:"mid"`
}

type BilibiliWebDynamicApiModuleDynamic struct {
	Major *struct {
		Opus *struct {
			Summary *BilibiliWebDynamicApiSummary `json:"summary"`
			Title   string                        `json:"title"`
			Pics    []*BilibiliWebDynamicApiPic   `json:"pics"`
		} `json:"opus"`
		Type string `json:"type"`
	} `json:"major"`
}

type BilibiliWebDynamicApiPic struct {
	Url    string  `json:"url"`
	Height int     `json:"height"`
	Width  int     `json:"width"`
	Size   float64 `json:"size"`
}

type BilibiliWebDynamicApiSummary struct {
	Text string `json:"text"`
}

func (resp *BilibiliWebDynamicApiResp) ToArtwork() (*dto.FetchedArtwork, error) {
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
	pictures := make([]*dto.FetchedPicture, 0, len(pics))
	for i, pic := range pics {
		pictures = append(pictures, &dto.FetchedPicture{
			Index:     uint(i),
			Original:  pic.Url,
			Width:     uint(pic.Width),
			Height:    uint(pic.Height),
			Thumbnail: pic.Url + "@1024w_1024h.jpg",
		})
	}
	title := opus.Title
	if title == "" {
		title = "bilibili/" + item.IdStr
	}
	artwork := &dto.FetchedArtwork{
		Title:       title,
		Description: summary.Text,
		SourceType:  shared.SourceTypeBilibili,
		SourceURL:   "https://t.bilibili.com/" + item.IdStr,
		R18:         false,
		Artist: &dto.FetchedArtist{
			Name:     author.Name,
			Username: author.Name,
			Type:     shared.SourceTypeBilibili,
			UID:      strconv.Itoa(author.Mid),
		},
		Pictures: pictures,
		Tags:     strutil.ExtractTagsFromText(summary.Text),
	}
	if err := checkArtworkField(artwork); err != nil {
		return nil, ErrInvalidURL
	}
	return artwork, nil
}

type BilibiliDesktopDynamicApiResp struct {
	Data *struct {
		Item *BilibiliDesktopDynamicApiItem `json:"item"`
	} `json:"data"`
	Message string `json:"message"`
	Code    int    `json:"code"`
}

type BilibiliDesktopDynamicApiItem struct {
	IdStr   string                             `json:"id_str"`
	Type    string                             `json:"type"`
	Modules []*BilibiliDesktopDynamicApiModule `json:"modules"`
}

type BilibiliDesktopDynamicApiModule struct {
	ModuleAuthor  *BilibiliDesktopDynamicApiAuthor  `json:"module_author,omitempty"`
	ModuleDesc    *BilibiliDesktopDynamicApiDesc    `json:"module_desc,omitempty"`
	ModuleDynamic *BilibiliDesktopDynamicApiDynamic `json:"module_dynamic,omitempty"`
	ModuleType    string                            `json:"module_type"`
}

type BilibiliDesktopDynamicApiAuthor struct {
	User *struct {
		Name string `json:"name"`
		Mid  int    `json:"mid"`
	} `json:"user"`
}

type BilibiliDesktopDynamicApiDesc struct {
	Text          string                                   `json:"text"`
	RichTextNodes []*BilibiliDesktopDynamicApiRichTextNode `json:"rich_text_nodes"`
}

type BilibiliDesktopDynamicApiRichTextNode struct {
	Emoji    *BilibiliDesktopDynamicApiEmoji `json:"emoji,omitempty"`
	Type     string                          `json:"type"`
	OrigText string                          `json:"orig_text"`
	Text     string                          `json:"text"`
}

type BilibiliDesktopDynamicApiEmoji struct {
	IconUrl string `json:"icon_url"`
	Text    string `json:"text"`
	Size    int    `json:"size"`
	Type    int    `json:"type"`
}

type BilibiliDesktopDynamicApiDynamic struct {
	DynDraw *struct {
		Items []*BilibiliDesktopDynamicApiDrawItem `json:"items"`
		Id    int                                  `json:"id"`
	} `json:"dyn_draw"`
	Type string `json:"type"`
}

type BilibiliDesktopDynamicApiDrawItem struct {
	Src    string   `json:"src"`
	Tags   []string `json:"tags"`
	Height int      `json:"height"`
	Width  int      `json:"width"`
	Size   float64  `json:"size"`
}

func (resp *BilibiliDesktopDynamicApiResp) ToArtwork() (*dto.FetchedArtwork, error) {
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
	if item.Type != "DYNAMIC_TYPE_DRAW" {
		return nil, ErrInvalidURL
	}
	modules := item.Modules
	if len(modules) == 0 {
		return nil, ErrInvalidURL
	}

	artwork := &dto.FetchedArtwork{
		SourceType: shared.SourceTypeBilibili,
		Title:      "bilibili/" + item.IdStr,
		SourceURL:  "https://t.bilibili.com/" + item.IdStr,
	}
	for _, module := range modules {
		if module == nil {
			continue
		}
		switch module.ModuleType {
		case "MODULE_TYPE_AUTHOR":
			if module.ModuleAuthor == nil || module.ModuleAuthor.User == nil {
				return nil, ErrInvalidURL
			}
			author := module.ModuleAuthor.User
			if author.Name == "" || author.Mid == 0 {
				return nil, ErrInvalidURL
			}
			artwork.Artist = &dto.FetchedArtist{
				Name:     author.Name,
				Username: author.Name,
				Type:     shared.SourceTypeBilibili,
				UID:      strconv.Itoa(author.Mid),
			}
		case "MODULE_TYPE_DESC":
			if module.ModuleDesc == nil {
				return nil, ErrInvalidURL
			}
			artwork.Description = module.ModuleDesc.Text
			artwork.Tags = strutil.ExtractTagsFromText(module.ModuleDesc.Text)
		case "MODULE_TYPE_DYNAMIC":
			if module.ModuleDynamic == nil {
				return nil, ErrInvalidURL
			}
			dynamic := module.ModuleDynamic
			if dynamic.DynDraw == nil || len(dynamic.DynDraw.Items) == 0 {
				return nil, ErrInvalidURL
			}
			pictures := make([]*dto.FetchedPicture, 0, len(dynamic.DynDraw.Items))
			for i, item := range dynamic.DynDraw.Items {
				if item.Src == "" {
					return nil, ErrInvalidURL
				}
				pictures = append(pictures, &dto.FetchedPicture{
					Index:     uint(i),
					Original:  item.Src,
					Width:     uint(item.Width),
					Height:    uint(item.Height),
					Thumbnail: item.Src + "@1024w_1024h.jpg",
				})
			}
			artwork.Pictures = pictures
		default:
			continue
		}
	}
	if err := checkArtworkField(artwork); err != nil {
		return nil, ErrInvalidURL
	}
	return artwork, nil
}

func checkArtworkField(artwork *dto.FetchedArtwork) error {
	if artwork.SourceType != shared.SourceTypeBilibili {
		return fmt.Errorf("%w: %v", ErrInvalidArtwork, artwork.SourceType)
	}
	if artwork.SourceURL == "" || artwork.Title == "" {
		return fmt.Errorf("%w: %v", ErrInvalidArtwork, artwork.SourceURL)
	}
	if artwork.Artist == nil || artwork.Artist.Name == "" || artwork.Artist.UID == "" {
		return fmt.Errorf("%w: %v", ErrInvalidArtwork, artwork.SourceURL)
	}
	if len(artwork.Pictures) == 0 {
		return fmt.Errorf("%w: %v", ErrInvalidArtwork, artwork.SourceURL)
	}
	for _, picture := range artwork.Pictures {
		if picture.Original == "" || picture.Thumbnail == "" {
			return fmt.Errorf("%w: %v", ErrInvalidArtwork, artwork.SourceURL)
		}
	}
	return nil
}

var ErrInvalidArtwork = errors.New("invalid artwork")
