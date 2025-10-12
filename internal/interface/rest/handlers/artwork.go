package handlers

import (
	"github.com/gofiber/fiber/v3"
	"github.com/krau/ManyACG/internal/infra/config/runtimecfg"
	"github.com/krau/ManyACG/internal/interface/rest/common"
	"github.com/krau/ManyACG/internal/interface/rest/utils"
	"github.com/krau/ManyACG/internal/model/entity"
	"github.com/krau/ManyACG/internal/model/query"
	"github.com/krau/ManyACG/internal/service"
	"github.com/krau/ManyACG/internal/shared"
	"github.com/krau/ManyACG/pkg/log"
	"github.com/krau/ManyACG/pkg/objectuuid"
	"github.com/krau/ManyACG/pkg/strutil"
)

type GetRandomArtworksRequest struct {
	R18   int `query:"r18,default=0" form:"r18,default=0" json:"r18" validate:"gte=0,lte=2" message:"R18 must be 0 (no R18), 1 (only R18) or 2 (both)"`
	Limit int `query:"limit,default=1" form:"limit,default=1" json:"limit" validate:"lte=200" message:"Limit must be between 1 and 200"`
	// Simple bool `form:"simple,default=false" json:"simple"` // deprecated
}

type ArtworkResponseItem struct {
	ID          string             `json:"id"`
	CreatedAt   string             `json:"created_at"`
	Title       string             `json:"title"`
	Description string             `json:"description"`
	SourceURL   string             `json:"source_url"`
	R18         bool               `json:"r18"`
	LikeCount   uint               `json:"like_count"`
	Tags        []string           `json:"tags"`
	Artist      *ArtistResponse    `json:"artist"`
	SourceType  shared.SourceType  `json:"source_type"`
	Pictures    []*PictureResponse `json:"pictures"`
}

type ArtistResponse struct {
	ID       string            `json:"id" bson:"_id"`
	Name     string            `json:"name" bson:"name"`
	Type     shared.SourceType `json:"type" bson:"type"`
	UID      string            `json:"uid" bson:"uid"`
	Username string            `json:"username" bson:"username"`
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

func artworksResponseFromEntity(artworks []*entity.Artwork, cfg runtimecfg.RestConfig, serv *service.Service) []*ArtworkResponseItem {
	resp := make([]*ArtworkResponseItem, 0, len(artworks))
	for _, art := range artworks {
		pics := make([]*PictureResponse, 0, len(art.Pictures))
		for _, pic := range art.Pictures {
			thumb, regular := utils.GetPictureResponseUrl(pic, cfg)
			pics = append(pics, &PictureResponse{
				ID:        pic.ID.Hex(),
				Width:     pic.Width,
				Height:    pic.Height,
				Index:     pic.OrderIndex,
				Hash:      pic.Phash,
				ThumbHash: pic.ThumbHash,
				FileName:  serv.PrettyFileName(art, pic),
				MessageID: pic.GetTelegramInfo().MessageID,
				Thumbnail: thumb,
				Regular:   regular,
			})
		}
		resp = append(resp, &ArtworkResponseItem{
			ID:          art.ID.Hex(),
			CreatedAt:   art.CreatedAt.Format("2006-01-02 15:04:05"),
			Title:       art.Title,
			Description: art.Description,
			SourceURL:   art.SourceURL,
			R18:         art.R18,
			LikeCount:   art.LikeCount,
			Tags:        art.GetTags(),
			Artist: &ArtistResponse{
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

func RandomArtworks(ctx fiber.Ctx) error {
	serv := common.MustGetState[*service.Service](ctx, common.StateKeyService)
	req := new(GetRandomArtworksRequest)
	if err := ctx.Bind().All(req); err != nil {
		return err
	}
	if req.Limit <= 0 {
		req.Limit = 1
	}
	artworks, err := serv.QueryArtworks(ctx, query.ArtworksDB{
		ArtworksFilter: query.ArtworksFilter{
			R18: shared.R18TypeFromInt(req.R18),
		},
		Random: true,
		Paginate: query.Paginate{
			Limit: req.Limit,
		},
	})
	if err != nil {
		return err
	}
	cfg := common.MustGetState[runtimecfg.RestConfig](ctx, common.StateKeyConfig)
	resp := artworksResponseFromEntity(artworks, cfg, serv)
	if len(resp) == 0 {
		return common.NewError(fiber.StatusNotFound, "no artworks found")
	}
	return ctx.JSON(common.NewSuccess(resp))
}

type GetArtworkListRequest struct {
	R18      int    `query:"r18" form:"r18" json:"r18" validate:"gte=0,lte=2" message:"R18 must be 0 (no R18), 1 (only R18) or 2 (both)"`
	ArtistID string `query:"artist_id" form:"artist_id" json:"artist_id" validate:"omitempty,objectid" message:"ArtistID must be a valid ObjectID"`
	Tag      string `query:"tag" form:"tag" json:"tag"`
	Keyword  string `query:"keyword" form:"keyword" json:"keyword" validate:"max=100" message:"Keyword max length is 100"`
	Page     int64  `query:"page" form:"page" json:"page"`
	PageSize int64  `query:"page_size" form:"page_size" json:"page_size" validate:"gte=0,lte=200" message:"PageSize must be between 1 and 200"`
	// Hybrid        bool   `query:"hybrid" form:"hybrid" json:"hybrid"`
	// SimilarTarget string `query:"similar_target" form:"similar_target" json:"similar_target"`
	// Simple        bool   `query:"simple" form:"simple" json:"simple"`
}

func ListArtworks(ctx fiber.Ctx) error {
	req := new(GetArtworkListRequest)
	if err := ctx.Bind().All(req); err != nil {
		return err
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}
	if req.Page <= 0 {
		req.Page = 1
	}
	var artistID objectuuid.ObjectUUID
	if req.ArtistID != "" {
		parsed, err := objectuuid.FromObjectIDHex(req.ArtistID)
		if err != nil {
			return err // should not happen due to validation
		}
		artistID = parsed
	}
	serv := common.MustGetState[*service.Service](ctx, common.StateKeyService)
	var tagId objectuuid.ObjectUUID
	if req.Tag != "" {
		tag, err := serv.GetTagByNameWithAlias(ctx, req.Tag)
		if err != nil {
			return common.NewError(fiber.StatusNotFound, "tag not found")
		}
		tagId = tag.ID
	}

	keywords := strutil.ParseTo2DArray(req.Keyword, ",", "|")
	dbQuery := query.ArtworksDB{
		ArtworksFilter: query.ArtworksFilter{
			R18: shared.R18TypeFromInt(req.R18),
		},
		Paginate: query.Paginate{
			Limit:  int(req.PageSize),
			Offset: int((req.Page - 1) * req.PageSize),
		},
	}
	if artistID != objectuuid.Nil {
		dbQuery.ArtistID = artistID
	}
	if tagId != objectuuid.Nil {
		dbQuery.Tags = [][]objectuuid.ObjectUUID{{tagId}}
	}
	if len(keywords) > 0 {
		dbQuery.Keywords = keywords
	}
	log.Debug("listing artworks", "query", dbQuery)
	artworks, err := serv.QueryArtworks(ctx, dbQuery)
	if err != nil {
		return err
	}
	cfg := common.MustGetState[runtimecfg.RestConfig](ctx, common.StateKeyConfig)
	resp := artworksResponseFromEntity(artworks, cfg, serv)
	if len(resp) == 0 {
		return common.NewError(fiber.StatusNotFound, "no artworks found")
	}
	return ctx.JSON(common.NewSuccess(resp))
}
