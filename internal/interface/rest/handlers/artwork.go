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
	"github.com/krau/ManyACG/internal/shared/errs"
	"github.com/krau/ManyACG/pkg/objectuuid"
	"github.com/krau/ManyACG/pkg/strutil"
)

type RequestRandomArtworks struct {
	R18   int `query:"r18,default=0" form:"r18,default=0" json:"r18" validate:"gte=0,lte=2" message:"r18 must be 0 (no R18), 1 (only R18) or 2 (both)"`
	Limit int `query:"limit,default=1" form:"limit,default=1" json:"limit" validate:"lte=200" message:"limit must be between 1 and 200"`
	// Simple bool `form:"simple,default=false" json:"simple"` // deprecated
}

type ResponseArtworkItem struct {
	ID          string             `json:"id"`
	CreatedAt   string             `json:"created_at"`
	Title       string             `json:"title"`
	Description string             `json:"description"`
	SourceURL   string             `json:"source_url"`
	R18         bool               `json:"r18"`
	LikeCount   uint               `json:"like_count"`
	Tags        []string           `json:"tags"`
	Artist      *ResponseArtist    `json:"artist"`
	SourceType  shared.SourceType  `json:"source_type"`
	Pictures    []*ResponsePicture `json:"pictures"`
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

func artworksResponseFromEntity(ctx fiber.Ctx, artworks []*entity.Artwork, cfg runtimecfg.RestConfig, serv *service.Service) []*ResponseArtworkItem {
	resp := make([]*ResponseArtworkItem, 0, len(artworks))
	for _, art := range artworks {
		pics := make([]*ResponsePicture, 0, len(art.Pictures))
		for _, pic := range art.Pictures {
			thumb, regular := utils.PictureResponseUrl(ctx, pic, cfg)
			pics = append(pics, &ResponsePicture{
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
		resp = append(resp, &ResponseArtworkItem{
			ID:          art.ID.Hex(),
			CreatedAt:   art.CreatedAt.Format("2006-01-02 15:04:05"),
			Title:       art.Title,
			Description: art.Description,
			SourceURL:   art.SourceURL,
			R18:         art.R18,
			LikeCount:   art.LikeCount,
			Tags:        art.GetTags(),
			Artist: &ResponseArtist{
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

func HandleRandomArtworks(ctx fiber.Ctx) error {
	serv := common.MustGetState[*service.Service](ctx, common.StateKeyService)
	req := new(RequestRandomArtworks)
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
	resp := artworksResponseFromEntity(ctx, artworks, cfg, serv)
	if len(resp) == 0 {
		return common.NewError(fiber.StatusNotFound, "no artworks found")
	}
	return ctx.JSON(common.NewSuccess(resp))
}

type RequestListArtworks struct {
	R18           int    `query:"r18" form:"r18" json:"r18" validate:"gte=0,lte=2" message:"r18 must be 0 (no R18), 1 (only R18) or 2 (both)"`
	ArtistID      string `query:"artist_id" form:"artist_id" json:"artist_id" validate:"omitempty,objectid" message:"artist_id must be a valid ObjectID"`
	Tag           string `query:"tag" form:"tag" json:"tag"`
	Keyword       string `query:"keyword" form:"keyword" json:"keyword" validate:"max=100" message:"keyword max length is 100"`
	Page          int64  `query:"page" form:"page" json:"page"`
	PageSize      int64  `query:"page_size" form:"page_size" json:"page_size" validate:"omitempty,gte=0,lte=200" message:"page_size must be between 1 and 200"`
	Hybrid        bool   `query:"hybrid" form:"hybrid" json:"hybrid"`
	SimilarTarget string `query:"similar_target" form:"similar_target" json:"similar_target" validate:"omitempty,objectid" message:"similar_target must be a valid ObjectID"`
	// Simple        bool   `query:"simple" form:"simple" json:"simple"`
}

func HandleListArtworks(ctx fiber.Ctx) error {
	req := new(RequestListArtworks)
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
			return err
		}
		artistID = parsed
	}

	serv := common.MustGetState[*service.Service](ctx, common.StateKeyService)
	cfg := common.MustGetState[runtimecfg.RestConfig](ctx, common.StateKeyConfig)

	var artworks []*entity.Artwork
	var err error

	if req.SimilarTarget != "" {
		targetId, err := objectuuid.FromObjectIDHex(req.SimilarTarget)
		if err != nil {
			return err
		}
		artworks, err = serv.FindSimilarArtworks(ctx, &query.ArtworkSimilar{
			ArtworkID: targetId,
			R18:       shared.R18TypeFromInt(req.R18),
			Paginate: query.Paginate{
				Limit:  int(req.PageSize),
				Offset: int((req.Page - 1) * req.PageSize),
			},
		})
		if err != nil {
			return err
		}
	} else if req.Hybrid {
		artworks, err = serv.SearchArtworks(ctx, &query.ArtworkSearch{
			Hybrid:              req.Hybrid,
			HybridSemanticRatio: 0.8,
			R18:                 shared.R18TypeFromInt(req.R18),
			Query:               req.Keyword,
			Paginate: query.Paginate{
				Limit:  int(req.PageSize),
				Offset: int((req.Page - 1) * req.PageSize),
			},
		})
		if err != nil {
			return err
		}
	} else {
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
		artworks, err = serv.QueryArtworks(ctx, dbQuery)
	}
	if err != nil {
		return err
	}
	if len(artworks) == 0 {
		return common.NewError(fiber.StatusNotFound, "no artworks found")
	}
	resp := artworksResponseFromEntity(ctx, artworks, cfg, serv)
	return ctx.JSON(common.NewSuccess(resp))
}

type RequestCountArtwork struct {
	R18 int `query:"r18" form:"r18" json:"r18" validate:"omitempty,gte=0,lte=2" message:"r18 must be 0 (no R18), 1 (only R18) or 2 (both)"`
}

func HandleCountArtwork(ctx fiber.Ctx) error {
	serv := common.MustGetState[*service.Service](ctx, common.StateKeyService)
	req := new(RequestCountArtwork)
	if err := ctx.Bind().All(req); err != nil {
		return err
	}
	count, err := serv.CountArtworks(ctx, shared.R18TypeFromInt(req.R18))
	if err != nil {
		return err
	}
	return ctx.JSON(common.NewSuccess(count))
}

type RequestFetchArtwork struct {
	URL string `query:"url" form:"url" json:"url" validate:"required,url" message:"url is required and must be a valid url"`
	// NoCache bool   `query:"no_cache" form:"no_cache" json:"no_cache"` // deprecated
}

type ResponseFetchArtwork struct {
	CacheID     string                    `json:"cache_id"`
	Title       string                    `json:"title"`
	Description string                    `json:"description"`
	SourceURL   string                    `json:"source_url"`
	R18         bool                      `json:"r18"`
	Tags        []string                  `json:"tags"`
	Artist      *ResponseFetchedArtist    `json:"artist"`
	SourceType  shared.SourceType         `json:"source_type"`
	Pictures    []*ResponseFetchedPicture `json:"pictures"`
}

type ResponseFetchedArtist struct {
	Name     string `json:"name"`
	Username string `json:"username"`
	UID      string `json:"uid"`
}

type ResponseFetchedPicture struct {
	Width     uint   `json:"width"`
	Height    uint   `json:"height"`
	Index     uint   `json:"index"`
	Thumbnail string `json:"thumbnail"`
	Original  string `json:"original"`
	FileName  string `json:"file_name"`
}

func fetchArtworkResponse(cacheID string, art shared.ArtworkLike, serv *service.Service) *ResponseFetchArtwork {
	pics := make([]*ResponseFetchedPicture, 0, len(art.GetPictures()))
	for _, pic := range art.GetPictures() {
		width, height := pic.GetSize()
		pics = append(pics, &ResponseFetchedPicture{
			Width:     width,
			Height:    height,
			Index:     pic.GetIndex(),
			Thumbnail: pic.GetThumbnail(),
			Original:  pic.GetOriginal(),
			FileName:  serv.PrettyFileName(art, pic),
		})
	}
	return &ResponseFetchArtwork{
		CacheID:     cacheID,
		Title:       art.GetTitle(),
		Description: art.GetDescription(),
		SourceURL:   art.GetSourceURL(),
		R18:         art.GetR18(),
		Tags:        art.GetTags(),
		Artist: &ResponseFetchedArtist{
			Name:     art.GetArtist().GetName(),
			Username: art.GetArtist().GetUserName(),
			UID:      art.GetArtist().GetUID(),
		},
		SourceType: art.GetType(),
		Pictures:   pics,
	}
}

func HandleFetchArtwork(ctx fiber.Ctx) error {
	key := ctx.Get("X-API-KEY")
	if key == "" {
		return common.NewError(fiber.StatusUnauthorized, "api key is required")
	}
	serv := common.MustGetState[*service.Service](ctx, common.StateKeyService)
	keyEnt, err := serv.GetApiKeyByKey(ctx, key)
	if err != nil {
		return common.NewError(fiber.StatusUnauthorized, "invalid api key")
	}
	if !keyEnt.HasPermission(shared.PermissionFetchArtwork) {
		return common.NewError(fiber.StatusForbidden, "no permission to fetch artwork")
	}
	if !keyEnt.CanUse() {
		return common.NewError(fiber.StatusForbidden, "api key quota exceeded")
	}

	req := new(RequestFetchArtwork)
	if err := ctx.Bind().All(req); err != nil {
		return err
	}
	sourceUrl := serv.FindSourceURL(req.URL)
	if sourceUrl == "" {
		return common.NewError(fiber.StatusBadRequest, "no valid source url found")
	}
	artwork, err := serv.GetOrFetchCachedArtwork(ctx, sourceUrl)
	if err != nil {
		return err
	}
	return ctx.JSON(common.NewSuccess(fetchArtworkResponse(artwork.ID.Hex(), artwork.Artwork.Data(), serv)))
}

func HandleGetArtworkByID(ctx fiber.Ctx) error {
	artworkId := ctx.Params("id")
	if artworkId == "" {
		return common.NewError(fiber.StatusBadRequest, "artwork id is required")
	}
	artworkUUID, err := objectuuid.FromObjectIDHex(artworkId)
	if err != nil {
		return common.NewError(fiber.StatusBadRequest, "invalid artwork id")
	}
	serv := common.MustGetState[*service.Service](ctx, common.StateKeyService)
	artwork, err := serv.GetArtworkByID(ctx, artworkUUID)
	if err != nil {
		if err == errs.ErrRecordNotFound {
			return common.NewError(fiber.StatusNotFound, "artwork not found")
		}
		return err
	}
	cfg := common.MustGetState[runtimecfg.RestConfig](ctx, common.StateKeyConfig)
	resp := artworksResponseFromEntity(ctx, []*entity.Artwork{artwork}, cfg, serv)
	return ctx.JSON(common.NewSuccess(resp[0]))
}
