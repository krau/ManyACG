package kemono

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/duke-git/lancet/v2/slice"
	"github.com/krau/ManyACG/internal/model/dto"
	"github.com/krau/ManyACG/internal/shared"
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
		} `json:"file"` // 封面?
		Tags        []string               `json:"tags"`
		Attachments []KemonoPostAttachment `json:"attachments"`
	} `json:"post"`
	Previews []KemonoPreview `json:"previews"`
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

type KemonoPreview struct {
	Type   string `json:"type"`   // "thumbnail" or others?
	Server string `json:"server"` // cdn server, e.g. "https://n4.kemono.cr"
	Name   string `json:"name"`   // file name, 好像没啥用
	Path   string `json:"path"`   // file path, 和原图一致
}

var htmlRe = regexp.MustCompile("<[^>]+>")

func (resp *KemonoPostResp) ToArtwork() (*dto.FetchedArtwork, error) {
	postResp := resp.Post
	creatorResp, err := getAuthorProfile(postResp.Service, postResp.User)
	if err != nil {
		return nil, err
	}
	artist := &dto.FetchedArtist{
		Type:     shared.SourceTypeKemono,
		Name:     creatorResp.Name,
		Username: postResp.Service + "_" + creatorResp.PubilcID,
		UID:      creatorResp.ID,
	}
	picCdnMap := make(map[string]string)
	for _, preview := range resp.Previews {
		if preview.Type != "thumbnail" {
			continue
		}
		picCdnMap[preview.Path] = preview.Server
	}
	pictures := make([]*dto.FetchedPicture, 0)
	if isImage(postResp.File.Path) {
		thumbnailUrl, err := url.JoinPath(thumbnailsBase, postResp.File.Path)
		if err != nil {
			return nil, fmt.Errorf("failed to join thumbnail path: %w", err)
		}
		originalUrl, err := url.JoinPath(picCdnMap[postResp.File.Path], "data", postResp.File.Path)
		if err != nil {
			return nil, fmt.Errorf("failed to join original path: %w", err)
		}
		pictures = append(pictures, &dto.FetchedPicture{
			Index:     0,
			Thumbnail: thumbnailUrl,
			Original:  originalUrl,
		})
	}
	for i, attachment := range postResp.Attachments {
		if !isImage(attachment.Path) {
			continue
		}
		thumbnailUrl, err := url.JoinPath(thumbnailsBase, attachment.Path)
		if err != nil {
			return nil, fmt.Errorf("failed to join thumbnail path: %w", err)
		}
		originalUrl, err := url.JoinPath(picCdnMap[attachment.Path], "data", attachment.Path)
		if err != nil {
			return nil, fmt.Errorf("failed to join original path: %w", err)
		}
		pictures = append(pictures, &dto.FetchedPicture{
			Index:     uint(i + 1),
			Thumbnail: thumbnailUrl,
			Original:  originalUrl,
		})
	}
	if len(pictures) == 0 {
		return nil, ErrNotPicture
	}
	pictures = slice.UniqueByComparator(pictures, func(item, other *dto.FetchedPicture) bool {
		return item.Original == other.Original
	})
	for i, pic := range pictures {
		pic.Index = uint(i)
	}
	artwork := &dto.FetchedArtwork{
		Title:       postResp.Title,
		Description: htmlRe.ReplaceAllString(strings.ReplaceAll(postResp.Content, "<br/>", "\n"), ""),
		R18:         false,
		SourceType:  shared.SourceTypeKemono,
		SourceURL:   fmt.Sprintf("https://kemono.cr/%s/user/%s/post/%s", postResp.Service, postResp.User, postResp.ID),
		Artist:      artist,
		Tags:        postResp.Tags,
		Pictures:    pictures,
	}
	return artwork, nil
}

// Deprecated: 2025-08-01: kemono has changed its api and now uses a different domain.
// func (resp *KemonoPostResp) ToArtwork() (*types.Artwork, error) {
// 	postResp := resp.Post
// 	creatorResp, err := getAuthorProfile(postResp.Service, postResp.User)
// 	if err != nil {
// 		return nil, err
// 	}
// 	artist := &types.Artist{
// 		Type:     types.SourceTypeKemono,
// 		Name:     creatorResp.Name,
// 		Username: postResp.Service + "_" + creatorResp.PubilcID,
// 		UID:      creatorResp.ID,
// 	}

// 	workerCount := config.Get().Source.Kemono.Worker
// 	jobs := make(chan struct {
// 		path  string
// 		index int
// 	})
// 	results := make(chan pictureResult, len(postResp.Attachments)+1)

// 	var wg sync.WaitGroup
// 	for range workerCount {
// 		wg.Add(1)
// 		go func() {
// 			defer wg.Done()
// 			for job := range jobs {
// 				fileURL := cdnBaseURL + job.path
// 				thumbnailURL := thumbnailsBaseURL + job.path
// 				common.Logger.Tracef("request image: %s", fileURL)
// 				fileResp, err := reqClient.R().DisableAutoReadResponse().Get(fileURL)
// 				if err != nil {
// 					common.Logger.Warnf("Failed to fetch image: %v", err)
// 					results <- pictureResult{nil, job.index, err}
// 					continue
// 				}

// 				if fileResp.StatusCode != http.StatusOK {
// 					common.Logger.Warnf("Failed to fetch image: status code: %d", fileResp.StatusCode)
// 					fileResp.Body.Close()
// 					results <- pictureResult{nil, job.index, fmt.Errorf("status code: %d", fileResp.StatusCode)}
// 					continue
// 				}
// 				if fileResp.Response != nil &&
// 					fileResp.Response.Request != nil &&
// 					fileResp.Response.Request.Response != nil &&
// 					fileResp.Response.Request.Response.Header != nil &&
// 					fileResp.Response.Request.Response.Header.Get("Location") != "" {
// 					fileURL = fileResp.Response.Request.Response.Header.Get("Location")
// 				}

// 				fileResp.Body.Close()
// 				common.Logger.Tracef("image fetched: %s", fileURL)

// 				results <- pictureResult{
// 					picture: &types.Picture{
// 						Index:     uint(job.index),
// 						Thumbnail: thumbnailURL,
// 						Original:  fileURL,
// 						Width:     0,
// 						Height:    0,
// 					},
// 					index: job.index,
// 					err:   nil,
// 				}
// 			}
// 		}()
// 	}

// 	go func() {
// 		if isImage(postResp.File.Path) {
// 			jobs <- struct {
// 				path  string
// 				index int
// 			}{postResp.File.Path, 0}
// 		}

// 		for i, attachment := range postResp.Attachments {
// 			if !isImage(attachment.Path) {
// 				continue
// 			}
// 			jobs <- struct {
// 				path  string
// 				index int
// 			}{attachment.Path, i + 1}
// 		}
// 		close(jobs)
// 	}()

// 	go func() {
// 		wg.Wait()
// 		close(results)
// 	}()

// 	// 排序
// 	pictureMap := make(map[int]*types.Picture)
// 	var maxIndex int
// 	for result := range results {
// 		if result.err != nil {
// 			common.Logger.Warnf("Failed to process image: %v", result.err)
// 			continue
// 		}
// 		pictureMap[result.index] = result.picture
// 		if result.index > maxIndex {
// 			maxIndex = result.index
// 		}
// 	}

// 	pictures := make([]*types.Picture, 0, len(pictureMap))
// 	for i := 0; i <= maxIndex; i++ {
// 		if pic, ok := pictureMap[i]; ok {
// 			pictures = append(pictures, pic)
// 		}
// 	}
// 	pictures = slice.UniqueByComparator(pictures, func(item, other *types.Picture) bool {
// 		return strings.EqualFold(item.Original, other.Original)
// 	})
// 	for i, pic := range pictures {
// 		pic.Index = uint(i)
// 	}

// 	if len(pictures) == 0 {
// 		return nil, ErrInvalidKemonoPostURL
// 	}

// 	artwork := &types.Artwork{
// 		Title:       postResp.Title,
// 		Description: htmlRe.ReplaceAllString(strings.ReplaceAll(postResp.Content, "<br/>", "\n"), ""),
// 		R18:         false,
// 		SourceType:  types.SourceTypeKemono,
// 		SourceURL:   fmt.Sprintf("https://kemono.cr/%s/user/%s/post/%s", postResp.Service, postResp.User, postResp.ID),
// 		Artist:      artist,
// 		Tags:        nil,
// 		Pictures:    pictures,
// 	}
// 	return artwork, nil
// }
