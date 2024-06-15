package danbooru

import (
	"ManyACG/types"
	"strconv"
	"strings"
)

type DanbooruSuccessJsonResp struct {
	ID           int    `json:"id"`
	ImageWidth   int    `json:"image_width"`
	ImageHeight  int    `json:"image_height"`
	TagString    string `json:"tag_string"`
	FileURL      string `json:"file_url"`
	LargeFileURL string `json:"large_file_url"`
}

type DanbooruFailJsonResp struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
	Message string `json:"message"`
}

func (resp *DanbooruSuccessJsonResp) ToArtwork() *types.Artwork {
	tags := strings.Split(resp.TagString, " ")
	pictures := make([]*types.Picture, 0)
	pictures = append(pictures, &types.Picture{
		Index:     0,
		Thumbnail: resp.LargeFileURL,
		Original:  resp.FileURL,
		Width:     uint(resp.ImageWidth),
		Height:    uint(resp.ImageHeight),
	})
	artwork := &types.Artwork{
		Title:       "Danbooru/" + strconv.Itoa(resp.ID),
		Description: "",
		R18:         false,
		SourceType:  types.SourceTypeDanbooru,
		SourceURL:   "https://danbooru.donmai.us/posts/" + strconv.Itoa(resp.ID),
		Artist:      fakeArtist,
		Tags:        tags,
		Pictures:    pictures,
	}
	return artwork
}
