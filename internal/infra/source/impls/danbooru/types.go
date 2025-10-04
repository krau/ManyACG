package danbooru

import (
	"strconv"
	"strings"

	"github.com/krau/ManyACG/internal/model/dto"
	"github.com/krau/ManyACG/internal/shared"
)

type DanbooruJsonResp struct {
	ID           int    `json:"id"`
	ImageWidth   int    `json:"image_width"`
	ImageHeight  int    `json:"image_height"`
	TagString    string `json:"tag_string"`
	FileURL      string `json:"file_url"`
	LargeFileURL string `json:"large_file_url"`
	DanbooruFailJsonResp
}

type DanbooruFailJsonResp struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
	Message string `json:"message"`
}

func (resp *DanbooruJsonResp) ToArtwork() *dto.FetchedArtwork {
	tags := strings.Split(resp.TagString, " ")
	pictures := make([]*dto.FetchedPicture, 0)
	pictures = append(pictures, &dto.FetchedPicture{
		Index:     0,
		Thumbnail: resp.LargeFileURL,
		Original:  resp.FileURL,
		Width:     uint(resp.ImageWidth),
		Height:    uint(resp.ImageHeight),
	})
	artwork := &dto.FetchedArtwork{
		Title:       "Danbooru/" + strconv.Itoa(resp.ID),
		Description: "",
		R18:         false,
		SourceType:  shared.SourceTypeDanbooru,
		SourceURL:   "https://danbooru.donmai.us/posts/" + strconv.Itoa(resp.ID),
		Artist:      fakeArtist,
		Tags:        tags,
		Pictures:    pictures,
	}
	return artwork
}
