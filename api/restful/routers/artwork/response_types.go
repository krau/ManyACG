package artwork

import "ManyACG/types"

type ArtworkResponse struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

type ArtworkResponseData struct {
	ID          string             `json:"id"`
	CreatedAt   string             `json:"created_at"`
	Title       string             `json:"title"`
	Description string             `json:"description"`
	SourceURL   string             `json:"source_url"`
	R18         bool               `json:"r18"`
	Tags        []string           `json:"tags"`
	Artist      *types.Artist      `json:"artist"`
	SourceType  types.SourceType   `json:"source_type"`
	Pictures    []*PictureResponse `json:"pictures"`
}

type PictureResponse struct {
	Width       uint               `json:"width"`
	Height      uint               `json:"height"`
	Index       uint               `json:"index"`
	Hash        string             `json:"hash"`
	BlurScore   float64            `json:"blur_score"`
	StorageInfo *types.StorageInfo `json:"storage_info"`
}

func ResponseFromArtwork(artwork *types.Artwork, isAuthorized bool) *ArtworkResponse {
	if isAuthorized {
		return &ArtworkResponse{
			Status:  200,
			Message: "Success",
			Data:    artwork,
		}
	}
	return &ArtworkResponse{
		Status:  200,
		Message: "Success",
		Data:    ResponseDataFromArtwork(artwork),
	}
}

func ResponseDataFromArtwork(artwork *types.Artwork) *ArtworkResponseData {
	pictures := make([]*PictureResponse, len(artwork.Pictures))
	for i, picture := range artwork.Pictures {
		pictures[i] = &PictureResponse{
			Width:       picture.Width,
			Height:      picture.Height,
			Index:       picture.Index,
			Hash:        picture.Hash,
			BlurScore:   picture.BlurScore,
			StorageInfo: picture.StorageInfo,
		}
	}
	return &ArtworkResponseData{
		ID:          artwork.ID,
		CreatedAt:   artwork.CreatedAt.Format("2006-01-02 15:04:05"),
		Title:       artwork.Title,
		Description: artwork.Description,
		SourceURL:   artwork.SourceURL,
		R18:         artwork.R18,
		Tags:        artwork.Tags,
		Artist:      artwork.Artist,
		SourceType:  artwork.SourceType,
		Pictures:    pictures,
	}
}

func ResponseFromArtworks(artworks []*types.Artwork, isAuthorized bool) *ArtworkResponse {
	if isAuthorized {
		return &ArtworkResponse{
			Status:  200,
			Message: "Success",
			Data:    artworks,
		}
	}
	responses := make([]*ArtworkResponseData, 0, len(artworks))
	for _, artwork := range artworks {
		responses = append(responses, ResponseDataFromArtwork(artwork))
	}
	return &ArtworkResponse{
		Status:  200,
		Message: "Success",
		Data:    responses,
	}
}