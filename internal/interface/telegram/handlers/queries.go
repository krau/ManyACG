package handlers

import (
	"strings"

	"github.com/krau/ManyACG/internal/common/restclient"
	"github.com/krau/ManyACG/internal/model/entity"
	"github.com/krau/ManyACG/internal/model/query"
	"github.com/krau/ManyACG/internal/service"
	"github.com/krau/ManyACG/internal/shared"
	"github.com/mymmrac/telego/telegohandler"
)

const maxQueryKeywordRunes = 200

// 远程只选作品, 本节点实体保留发图和回写所需信息。
func queryRandomArtworks(ctx *telegohandler.Context, serv *service.Service, params service.ArtworkQueryParams) ([]*entity.Artwork, error) {
	params.Keyword = truncateQueryKeyword(params.Keyword)
	client := restClient(ctx)
	if !client.Enabled() {
		return serv.QueryArtworksFiltered(ctx, params)
	}
	ids, err := client.RandomArtworks(ctx, restclient.ArtworkQuery{
		R18:      params.R18,
		Limit:    params.Limit,
		Keyword:  params.Keyword,
		ArtistID: params.ArtistID,
		Tag:      params.Tag,
	})
	if err != nil {
		return nil, err
	}
	return serv.GetArtworksByIDs(ctx, ids)
}

func searchHybridArtworks(ctx *telegohandler.Context, serv *service.Service, keyword string, ratio float64, limit int) ([]*entity.Artwork, error) {
	keyword = truncateQueryKeyword(keyword)
	client := restClient(ctx)
	if !client.Enabled() {
		return serv.SearchArtworks(ctx, &query.ArtworkSearch{
			Hybrid:              true,
			HybridSemanticRatio: ratio,
			Query:               keyword,
			Limit:               limit,
			R18:                 shared.R18TypeAll,
		})
	}
	ids, err := client.ListArtworks(ctx, restclient.ArtworkQuery{
		R18:           int(shared.R18TypeAll),
		Limit:         limit,
		Keyword:       keyword,
		Hybrid:        true,
		SemanticRatio: ratio,
	})
	if err != nil {
		return nil, err
	}
	return serv.GetArtworksByIDs(ctx, ids)
}

// keyword 中逗号连接 OR, 竖线分隔 AND 组。
func keywordParam(groups [][]string) string {
	rows := make([]string, 0, len(groups))
	for _, group := range groups {
		terms := make([]string, 0, len(group))
		for _, term := range group {
			terms = append(terms, quoteKeywordTerm(term))
		}
		rows = append(rows, strings.Join(terms, ","))
	}
	return strings.Join(rows, "|")
}

func quoteKeywordTerm(term string) string {
	if !strings.ContainsAny(term, ",|\"'") {
		return term
	}
	return `"` + strings.NewReplacer(`"`, "", `'`, "").Replace(term) + `"`
}

func truncateQueryKeyword(keyword string) string {
	count := 0
	for i := range keyword {
		if count == maxQueryKeywordRunes {
			return keyword[:i]
		}
		count++
	}
	return keyword
}
