package artwork

import (
	"net/http"

	"github.com/krau/ManyACG/common"
	"github.com/krau/ManyACG/sources"
	"github.com/krau/ManyACG/types"
)

type ArtworkResponseData struct {
	ID          string             `json:"id"`
	CreatedAt   string             `json:"created_at"`
	Title       string             `json:"title"`
	Description string             `json:"description"`
	SourceURL   string             `json:"source_url"`
	R18         bool               `json:"r18"`
	LikeCount   uint               `json:"like_count"`
	Tags        []string           `json:"tags"`
	Artist      *types.Artist      `json:"artist"`
	SourceType  types.SourceType   `json:"source_type"`
	Pictures    []*PictureResponse `json:"pictures"`
}

type PictureResponse struct {
	ID        string `json:"id"`
	Width     uint   `json:"width"`
	Height    uint   `json:"height"`
	Index     uint   `json:"index"`
	Hash      string `json:"hash"`
	ThumbHash string `json:"thumb_hash"`
	FileName  string `json:"file_name"`
	MessageID int    `json:"message_id"`
	Thumbnail string `json:"thumbnail"`
	Regular   string `json:"regular"`
}

func ResponseFromArtwork(artwork *types.Artwork, isAuthorized bool) *common.RestfulCommonResponse[any] {
	if isAuthorized {
		return &common.RestfulCommonResponse[any]{
			Status:  http.StatusOK,
			Message: "Success",
			Data:    artwork,
		}
	}
	return &common.RestfulCommonResponse[any]{
		Status:  http.StatusOK,
		Message: "Success",
		Data:    ResponseDataFromArtwork(artwork),
	}
}

func ResponseDataFromArtwork(artwork *types.Artwork) *ArtworkResponseData {
	pictures := make([]*PictureResponse, len(artwork.Pictures))
	for i, picture := range artwork.Pictures {
		var thumbnail, regular string
		if picture.StorageInfo == nil || picture.StorageInfo.Thumb == nil {
			thumbnail = picture.Thumbnail
		} else {
			picThumbUrl := common.ApplyApiStoragePathRule(picture.StorageInfo.Thumb)
			if picThumbUrl == "" || picThumbUrl == picture.StorageInfo.Thumb.Path {
				thumbnail = picture.Thumbnail
			} else {
				thumbnail = picThumbUrl
			}
		}
		if picture.StorageInfo == nil || picture.StorageInfo.Regular == nil {
			regular = picture.Thumbnail
		} else {
			picRegularUrl := common.ApplyApiStoragePathRule(picture.StorageInfo.Regular)
			if picRegularUrl == "" || picRegularUrl == picture.StorageInfo.Regular.Path {
				regular = picture.Thumbnail
			} else {
				regular = picRegularUrl
			}
		}
		pictures[i] = &PictureResponse{
			ID:     picture.ID,
			Width:  picture.Width,
			Height: picture.Height,
			Index:  picture.Index,
			Hash:   picture.Hash,
			FileName: func() string {
				fileName, err := sources.GetFileName(artwork, picture)
				if err != nil {
					return picture.GetFileName()
				}
				return fileName
			}(),
			ThumbHash: picture.ThumbHash,
			MessageID: picture.TelegramInfo.MessageID,
			Thumbnail: thumbnail,
			Regular:   regular,
		}
	}
	return &ArtworkResponseData{
		ID:          artwork.ID,
		CreatedAt:   artwork.CreatedAt.Format("2006-01-02 15:04:05"),
		Title:       artwork.Title,
		Description: artwork.Description,
		SourceURL:   artwork.SourceURL,
		R18:         artwork.R18,
		LikeCount:   artwork.LikeCount,
		Tags:        artwork.Tags,
		Artist:      artwork.Artist,
		SourceType:  artwork.SourceType,
		Pictures:    pictures,
	}
}

func ResponseFromArtworks(artworks []*types.Artwork, isAuthorized bool) *common.RestfulCommonResponse[any] {
	if isAuthorized {
		return &common.RestfulCommonResponse[any]{
			Status:  http.StatusOK,
			Message: "Success",
			Data:    artworks,
		}
	}
	responses := make([]*ArtworkResponseData, 0, len(artworks))
	for _, artwork := range artworks {
		responses = append(responses, ResponseDataFromArtwork(artwork))
	}
	return &common.RestfulCommonResponse[any]{
		Status:  http.StatusOK,
		Message: "Success",
		Data:    responses,
	}
}

type FetchedArtworkResponseData struct {
	CacheID     string                    `json:"cache_id"`
	Title       string                    `json:"title"`
	Description string                    `json:"description"`
	SourceURL   string                    `json:"source_url"`
	R18         bool                      `json:"r18"`
	Tags        []string                  `json:"tags"`
	Artist      *FetchedArtistResponse    `json:"artist"`
	SourceType  types.SourceType          `json:"source_type"`
	Pictures    []*FetchedPictureResponse `json:"pictures"`
}

type FetchedArtistResponse struct {
	Name     string `json:"name"`
	Username string `json:"username"`
	UID      string `json:"uid"`
}

type FetchedPictureResponse struct {
	Width     uint   `json:"width"`
	Height    uint   `json:"height"`
	Index     uint   `json:"index"`
	Thumbnail string `json:"thumbnail"`
	Original  string `json:"original"`
	FileName  string `json:"file_name"`
}

func ResponseFromFetchedArtwork(artwork *types.Artwork, cacheID string) *common.RestfulCommonResponse[FetchedArtworkResponseData] {
	return &common.RestfulCommonResponse[FetchedArtworkResponseData]{
		Status:  http.StatusOK,
		Message: "Success",
		Data:    ResponseDataFromFetchedArtwork(artwork, cacheID),
	}
}

func ResponseDataFromFetchedArtwork(artwork *types.Artwork, cacheID string) FetchedArtworkResponseData {
	pictures := make([]*FetchedPictureResponse, 0, len(artwork.Pictures))
	for _, picture := range artwork.Pictures {
		pictures = append(pictures, &FetchedPictureResponse{
			Width:     picture.Width,
			Height:    picture.Height,
			Index:     picture.Index,
			Thumbnail: picture.Thumbnail,
			Original:  picture.Original,
			FileName: func() string {
				fileName, err := sources.GetFileName(artwork, picture)
				if err != nil {
					return picture.GetFileName()
				}
				return fileName
			}(),
		})
	}
	return FetchedArtworkResponseData{
		CacheID:     cacheID,
		Title:       artwork.Title,
		Description: artwork.Description,
		SourceURL:   artwork.SourceURL,
		R18:         artwork.R18,
		Tags:        artwork.Tags,
		Artist: &FetchedArtistResponse{
			Name:     artwork.Artist.Name,
			Username: artwork.Artist.Username,
			UID:      artwork.Artist.UID,
		},
		SourceType: artwork.SourceType,
		Pictures:   pictures,
	}
}
