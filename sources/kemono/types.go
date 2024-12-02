package kemono

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/krau/ManyACG/common"
	"github.com/krau/ManyACG/types"
)

type KemonoPostResp struct {
	Post struct {
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
	} `json:"post"`
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
	postResp := resp.Post
	creatorResp, err := getAuthorProfile(postResp.Service, postResp.User)
	if err != nil {
		return nil, err
	}
	artist := &types.Artist{
		Type:     types.SourceTypeKemono,
		Name:     creatorResp.Name,
		Username: postResp.Service + "_" + creatorResp.PubilcID,
		UID:      creatorResp.ID,
	}
	pictures := make([]*types.Picture, 0)
	if isImage(postResp.File.Path) {
		fileResp, err := reqClient.R().Get(cdnBaseURL + postResp.File.Path)
		if err == nil && fileResp.StatusCode == http.StatusOK {
			fileUrl := cdnBaseURL + postResp.File.Path
			if fileResp.Response != nil &&
				fileResp.Response.Request != nil &&
				fileResp.Response.Request.Response != nil &&
				fileResp.Response.Request.Response.Header != nil &&
				fileResp.Response.Request.Response.Header.Get("Location") != "" {
				fileUrl = fileResp.Response.Request.Response.Header.Get("Location")
			}
			pictures = append(pictures, &types.Picture{
				Index:     0,
				Thumbnail: thumbnailsBaseURL + postResp.File.Path,
				Original:  fileUrl,
				Width:     0,
				Height:    0,
			})
		}
		fileResp.Body.Close()
	}
	i := len(pictures)
	for _, attachment := range postResp.Attachments {
		if !isImage(attachment.Path) {
			continue
		}
		fileURL := cdnBaseURL + attachment.Path
		fileResp, err := reqClient.R().DisableAutoReadResponse().Get(fileURL)
		if err != nil {
			common.Logger.Warnf("get attachment %s failed: %s", fileURL, err)
			continue
		}
		if fileResp.StatusCode != http.StatusOK {
			common.Logger.Warnf("get attachment %s failed: %d", fileURL, fileResp.StatusCode)
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
		thumbnailURL := thumbnailsBaseURL + attachment.Path
		pictures = append(pictures, &types.Picture{
			Index:     uint(i),
			Thumbnail: thumbnailURL,
			Original:  fileURL,
			Width:     0,
			Height:    0,
		})
		i++
	}
	if len(pictures) == 0 {
		return nil, ErrInvalidKemonoPostURL
	}
	artwork := &types.Artwork{
		Title:       postResp.Title,
		Description: htmlRe.ReplaceAllString(strings.ReplaceAll(postResp.Content, "<br/>", "\n"), ""),
		R18:         false,
		SourceType:  types.SourceTypeKemono,
		SourceURL:   fmt.Sprintf("https://kemono.su/%s/user/%s/post/%s", postResp.Service, postResp.User, postResp.ID),
		Artist:      artist,
		Tags:        nil,
		Pictures:    pictures,
	}
	return artwork, nil
}
