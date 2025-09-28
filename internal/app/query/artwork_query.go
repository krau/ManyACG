package query

import (
	"context"

	"github.com/krau/ManyACG/internal/common/decorator"
	"github.com/krau/ManyACG/internal/shared"
)

// query a single artwork
type ArtworkQuery struct {
	ID        string
	SourceURL string
}

type ArtistInfo struct {
	shared.ArtistInfo
	ID string `json:"id"`
}

type ArtworkQueryResult struct {
	ID          string                `json:"id"`
	Title       string                `json:"title"`
	Description string                `json:"description"`
	R18         bool                  `json:"r18"`
	SourceType  shared.SourceType     `json:"source_type"`
	SourceURL   string                `json:"source_url"`
	Artist      *ArtistInfo           `json:"artist"`
	Tags        []string              `json:"tags"`
	CreatedAt   string                `json:"created_at"`
	Pictures    []*shared.PictureInfo `json:"pictures"`
}

type ArtworkSearchQuery struct {
	R18      shared.R18Type
	ArtistID string
	Tags     [][]string // AND of OR tags
	Keywords [][]string // AND of OR keywords
	Hybrid   bool       // use hybrid search
	Query    string     // fulltext search query
	Limit    int
	Offset   int
}

type ArtworkSearchQueryResult = []*ArtworkQueryResult

type ArtworkQueryHandler decorator.QueryHandler[ArtworkQuery, *ArtworkQueryResult]

type ArtworkQueryRepo interface {
	FindByID(ctx context.Context, id string) (*ArtworkQueryResult, error)
	FindByURL(ctx context.Context, url string) (*ArtworkQueryResult, error)
}

type ArtworkSearchQueryHandler decorator.QueryHandler[ArtworkSearchQuery, ArtworkSearchQueryResult]

type ArtworkSearchQueryRepo interface {
	List(ctx context.Context, query ArtworkSearchQuery) (ArtworkSearchQueryResult, error)
	Count(ctx context.Context, query ArtworkSearchQuery) (int, error)
}

type artworkQueryHandler struct {
	queryRepo ArtworkQueryRepo
}

type artworkSearchQueryHandler struct {
	queryRepo ArtworkSearchQueryRepo
}

// Handle implements ArtworkSearchQueryHandler.
func (a *artworkSearchQueryHandler) Handle(ctx context.Context, query ArtworkSearchQuery) (ArtworkSearchQueryResult, error) {
	panic("unimplemented")
}

func NewArtworkSearchQueryHandler(queryRepo ArtworkSearchQueryRepo) ArtworkSearchQueryHandler {
	return &artworkSearchQueryHandler{
		queryRepo: queryRepo,
	}
}

// Handle implements ArtworkQueryHandler.
func (a *artworkQueryHandler) Handle(ctx context.Context, query ArtworkQuery) (*ArtworkQueryResult, error) {
	panic("unimplemented")
}

func NewArtworkQueryHandler(queryRepo ArtworkQueryRepo) ArtworkQueryHandler {
	return &artworkQueryHandler{
		queryRepo: queryRepo,
	}
}
