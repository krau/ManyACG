package kemono

import (
	"ManyACG/types"
	"fmt"
	"strconv"
)

type KemonoPostResp struct {
	Error   string `json:"error"`
	ID      string `json:"id"`
	User    string `json:"user"`
	Service string `json:"service"`
	Title   string `json:"title"`
	Content string `json:"content"`
	File    struct {
		Name string `json:"name"`
		Path string `json:"path"`
	} `json:"file"`
	Attachments []KemonoPostAttachment `json:"attachments"`
}

type KemonoPostAttachment struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

type KemonoCreatorProfileResp struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Service  string `json:"service"`
	PubilcID string `json:"public_id"` // username
}

func (resp *KemonoPostResp) ToArtwork() (*types.Artwork, error) {
	creatorResp, err := getAuthorProfile(resp.Service, resp.User)
	if err != nil {
		return nil, err
	}
	creatorIdInt, err := strconv.Atoi(creatorResp.ID)
	if err != nil {
		return nil, err
	}
	artist := &types.Artist{
		Type:     types.SourceTypeKemono,
		Name:     creatorResp.Name,
		Username: resp.Service + "_" + creatorResp.PubilcID,
		UID:      creatorIdInt,
	}
	pictures := make([]*types.Picture, 0)
	if isImage(resp.File.Path) {
		pictures = append(pictures, &types.Picture{
			Index:     0,
			Thumbnail: cdnBaseURL + resp.File.Path,
			Original:  cdnBaseURL + resp.File.Path,
			Width:     0,
			Height:    0,
		})
	}
	for i, attachment := range resp.Attachments {
		if !isImage(attachment.Path) {
			continue
		}
		pictures = append(pictures, &types.Picture{
			Index:     uint(i + 1),
			Thumbnail: cdnBaseURL + attachment.Path,
			Original:  cdnBaseURL + attachment.Path,
			Width:     0,
			Height:    0,
		})
	}
	artwork := &types.Artwork{
		Title:       resp.Title,
		Description: resp.Content,
		R18:         false,
		SourceType:  types.SourceTypeKemono,
		SourceURL:   fmt.Sprintf("https://kemono.su/%s/user/%s/post/%s", resp.Service, resp.User, resp.ID),
		Artist:      artist,
		Tags:        nil,
		Pictures:    pictures,
	}
	return artwork, nil
}
