package handlers

import (
	"context"
	"hash"
	"strconv"
	"strings"
	"time"

	"github.com/cespare/xxhash/v2"
	"github.com/dgraph-io/ristretto/v2"
	"github.com/gofiber/fiber/v3"
	"github.com/krau/ManyACG/internal/infra/config/runtimecfg"
	"github.com/krau/ManyACG/internal/interface/rest/common"
	"github.com/krau/ManyACG/internal/model/entity"
	"github.com/krau/ManyACG/internal/model/query"
	"github.com/krau/ManyACG/internal/service"
	"github.com/krau/ManyACG/internal/shared"
	"github.com/krau/ManyACG/pkg/log"
	"github.com/krau/ManyACG/pkg/strutil"
	"github.com/unvgo/ouid"
)

type RequestListArtworks struct {
	R18           int    `query:"r18" form:"r18" json:"r18" validate:"gte=0,lte=2" message:"r18 must be 0 (no R18), 1 (only R18) or 2 (both)"`
	ArtistID      string `query:"artist_id" form:"artist_id" json:"artist_id" validate:"omitempty,objectid" message:"artist_id must be a valid ObjectID"`
	Tag           string `query:"tag" form:"tag" json:"tag"`
	Keyword       string `query:"keyword" form:"keyword" json:"keyword" validate:"max=50" message:"keyword max length is 50"`
	Page          int64  `query:"page" form:"page" json:"page"`
	PageSize      int64  `query:"page_size" form:"page_size" json:"page_size" validate:"omitempty,gte=0,lte=200" message:"page_size must be between 1 and 200"`
	Hybrid        bool   `query:"hybrid" form:"hybrid" json:"hybrid"`
	SimilarTarget string `query:"similar_target" form:"similar_target" json:"similar_target" validate:"omitempty,objectid" message:"similar_target must be a valid ObjectID"`
}

func (r *RequestListArtworks) ID() uint64 {
	var b strings.Builder
	b.WriteString("r18=")
	b.WriteString(strconv.Itoa(r.R18))
	if r.ArtistID != "" {
		b.WriteString("&aid=")
		b.WriteString(r.ArtistID)
	}
	if r.Tag != "" {
		b.WriteString("&tag=")
		b.WriteString(r.Tag)
	}
	if r.Keyword != "" {
		b.WriteString("&kw=")
		b.WriteString(r.Keyword)
	}
	if r.Page != 0 {
		b.WriteString("&p=")
		b.WriteString(strconv.FormatInt(r.Page, 10))
	}
	if r.PageSize != 0 {
		b.WriteString("&ps=")
		b.WriteString(strconv.FormatInt(r.PageSize, 10))
	}
	if r.Hybrid {
		b.WriteString("&hy=1")
	}
	if r.SimilarTarget != "" {
		b.WriteString("&st=")
		b.WriteString(r.SimilarTarget)
	}

	var h hash.Hash64 = xxhash.New()
	_, _ = h.Write([]byte(b.String()))
	return h.Sum64()
}

func GetHandleListArtworks(serv *service.Service, cfg runtimecfg.RestConfig) fiber.Handler {
	var cache *ristretto.Cache[uint64, common.Response[[]ResponseArtworkItem]]
	var err error
	if !cfg.Cache.Disable {
		cache, err = ristretto.NewCache(&ristretto.Config[uint64, common.Response[[]ResponseArtworkItem]]{
			NumCounters:        1e4,
			MaxCost:            1e5,
			BufferItems:        64,
			IgnoreInternalCost: true,
			Cost:               func(common.Response[[]ResponseArtworkItem]) int64 { return 1 },
			OnReject: func(item *ristretto.Item[common.Response[[]ResponseArtworkItem]]) {
				log.Warn("cache reject item", "key", item.Key)
			},
		})
		if err != nil {
			log.Fatalf("failed to create cache: %v", err)
		}
	}
	ttl := time.Duration(cfg.Cache.DefaultTTL) * time.Second

	return func(ctx fiber.Ctx) error {
		requestCtx := ctx.RequestCtx()
		req := new(RequestListArtworks)
		if err := ctx.Bind().All(req); err != nil {
			return err
		}
		cacheKey := req.ID()
		if cache != nil {
			if cached, found := cache.Get(cacheKey); found {
				return ctx.JSON(cached)
			}
		}

		artworks, err := listArtworks(requestCtx, serv, req)
		if err != nil {
			return err
		}
		if len(artworks) == 0 {
			return common.NewError(fiber.StatusNotFound, "no artworks found")
		}
		items := artworksResponseFromEntity(ctx, artworks, cfg, serv)
		resp := common.NewSuccess(items)
		if cache != nil {
			cache.SetWithTTL(cacheKey, *resp, 1, ttl)
		}
		return ctx.JSON(resp)
	}
}

func listArtworks(ctx context.Context, serv *service.Service, req *RequestListArtworks) ([]*entity.Artwork, error) {
	if req.PageSize <= 0 {
		req.PageSize = 20
	}
	if req.Page <= 0 {
		req.Page = 1
	}

	var artistID ouid.OUID
	if req.ArtistID != "" {
		parsed, err := ouid.FromObjectIDHex(req.ArtistID)
		if err != nil {
			return nil, err
		}
		artistID = parsed
	}

	if req.SimilarTarget != "" {
		targetID, err := ouid.FromObjectIDHex(req.SimilarTarget)
		if err != nil {
			return nil, err
		}
		return serv.FindSimilarArtworks(ctx, &query.ArtworkSimilar{
			ArtworkID: targetID,
			R18:       shared.R18TypeFromInt(req.R18),
			Paginate: query.Paginate{
				Limit:  int(req.PageSize),
				Offset: int((req.Page - 1) * req.PageSize),
			},
		})
	}

	if req.Hybrid {
		return serv.SearchArtworks(ctx, &query.ArtworkSearch{
			Hybrid:              true,
			HybridSemanticRatio: 0.8,
			R18:                 shared.R18TypeFromInt(req.R18),
			Query:               req.Keyword,
			Paginate: query.Paginate{
				Limit:  int(req.PageSize),
				Offset: int((req.Page - 1) * req.PageSize),
			},
		})
	}

	var tagID ouid.OUID
	if req.Tag != "" {
		tag, err := serv.GetTagByNameWithAlias(ctx, req.Tag)
		if err != nil {
			return nil, common.NewError(fiber.StatusNotFound, "tag not found")
		}
		tagID = tag.ID
	}

	keywords := strutil.ParseTo2DArray(req.Keyword, ",", "|")
	dbQuery := query.ArtworksDB{
		ArtworksFilter: query.ArtworksFilter{
			R18:        shared.R18TypeFromInt(req.R18),
			HasPicture: true,
		},
		Paginate: query.Paginate{
			Limit:  int(req.PageSize),
			Offset: int((req.Page - 1) * req.PageSize),
		},
	}
	if artistID != ouid.Nil {
		dbQuery.ArtistID = artistID
	}
	if tagID != ouid.Nil {
		dbQuery.Tags = [][]ouid.OUID{{tagID}}
	}
	if len(keywords) > 0 {
		dbQuery.Keywords = keywords
	}
	return serv.QueryArtworks(ctx, dbQuery)
}
