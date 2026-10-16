package handlers

import (
	"github.com/gofiber/fiber/v3"
	"github.com/krau/ManyACG/internal/infra/config/runtimecfg"
	"github.com/krau/ManyACG/internal/interface/rest/utils"
	"github.com/krau/ManyACG/internal/model/entity"
	"github.com/krau/ManyACG/internal/service"
	"github.com/krau/ManyACG/internal/shared"
)

type ResponseArtworkItem struct {
	ID          string            `json:"id"`
	CreatedAt   string            `json:"created_at"`
	Title       string            `json:"title"`
	Description string            `json:"description"`
	SourceURL   string            `json:"source_url"`
	R18         bool              `json:"r18"`
	LikeCount   uint              `json:"like_count"`
	Tags        []string          `json:"tags"`
	Artist      ResponseArtist    `json:"artist"`
	SourceType  shared.SourceType `json:"source_type"`
	Pictures    []ResponsePicture `json:"pictures"`
}

type ResponseArtist struct {
	ID       string            `json:"id" bson:"_id"`
	Name     string            `json:"name" bson:"name"`
	Type     shared.SourceType `json:"type" bson:"type"`
	UID      string            `json:"uid" bson:"uid"`
	Username string            `json:"username" bson:"username"`
}

type ResponsePicture struct {
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

func artworksResponseFromEntity(ctx fiber.Ctx, artworks []*entity.Artwork, cfg runtimecfg.RestConfig, serv *service.Service) []ResponseArtworkItem {
	resp := make([]ResponseArtworkItem, 0, len(artworks))
	for _, art := range artworks {
		if len(art.Pictures) == 0 {
			continue
		}
		pics := make([]ResponsePicture, 0, len(art.Pictures))
		for _, pic := range art.Pictures {
			thumb, regular := utils.PictureResponseUrl(ctx, pic, cfg)
			pics = append(pics, ResponsePicture{
				ID:        pic.ID.Hex(),
				Width:     pic.Width,
				Height:    pic.Height,
				Index:     pic.OrderIndex,
				Hash:      pic.Phash,
				ThumbHash: pic.ThumbHash,
				FileName:  serv.PrettyFileName(art, pic),
				MessageID: 0,
				Thumbnail: thumb,
				Regular:   regular,
			})
		}
		resp = append(resp, ResponseArtworkItem{
			ID:          art.ID.Hex(),
			CreatedAt:   art.CreatedAt.Format("2006-01-02 15:04:05"),
			Title:       art.Title,
			Description: art.Description,
			SourceURL:   art.SourceURL,
			R18:         art.R18,
			LikeCount:   art.LikeCount,
			Tags:        art.GetTags(),
			Artist: ResponseArtist{
				ID:       art.Artist.ID.Hex(),
				Name:     art.Artist.Name,
				Type:     art.Artist.Type,
				UID:      art.Artist.UID,
				Username: art.Artist.Username,
			},
			SourceType: art.SourceType,
			Pictures:   pics,
		})
	}
	return resp
}
