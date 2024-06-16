package kemono

import (
	. "ManyACG/logger"
	"ManyACG/types"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
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

var htmlRe = regexp.MustCompile("<[^>]+>")

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
		fileResp, err := reqClient.R().DisableAutoReadResponse().Get(cdnBaseURL + resp.File.Path)
		if err == nil && fileResp.StatusCode == http.StatusOK {
			pictures = append(pictures, &types.Picture{
				Index:     0,
				Thumbnail: cdnBaseURL + resp.File.Path,
				Original:  cdnBaseURL + resp.File.Path,
				Width:     0,
				Height:    0,
			})
		}
		fileResp.Body.Close()
	}
	i := len(pictures)
	for _, attachment := range resp.Attachments {
		if !isImage(attachment.Path) {
			continue
		}
		fileURL := cdnBaseURL + attachment.Path
		fileResp, err := reqClient.R().DisableAutoReadResponse().Get(fileURL)
		if err != nil {
			Logger.Warnf("get attachment %s failed: %s", fileURL, err)
			continue
		}
		if fileResp.StatusCode != http.StatusOK {
			Logger.Warnf("get attachment %s failed: %d", fileURL, fileResp.StatusCode)
			continue
		}
		fileResp.Body.Close()
		isDuplicate := false
		for _, picture := range pictures {
			if picture.Original == fileURL {
				isDuplicate = true
				break
			}
		}
		if isDuplicate {
			continue
		}
		pictures = append(pictures, &types.Picture{
			Index:     uint(i),
			Thumbnail: fileURL,
			Original:  fileURL,
			Width:     0,
			Height:    0,
		})
		i++
	}
	artwork := &types.Artwork{
		Title:       resp.Title,
		Description: htmlRe.ReplaceAllString(strings.ReplaceAll(resp.Content, "<br/>", "\n"), ""),
		R18:         false,
		SourceType:  types.SourceTypeKemono,
		SourceURL:   fmt.Sprintf("https://kemono.su/%s/user/%s/post/%s", resp.Service, resp.User, resp.ID),
		Artist:      artist,
		Tags:        nil,
		Pictures:    pictures,
	}
	return artwork, nil
}
