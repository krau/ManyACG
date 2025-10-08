package dto

import (
	"github.com/krau/ManyACG/pkg/objectuuid"
)

type ArtworkSearchDocument struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Artist      string   `json:"artist"`
	Description string   `json:"description"`
	Tags        []string `json:"tags"`
	R18         bool     `json:"r18"`
}

type ArtworkSearchResult struct {
	IDs []objectuuid.ObjectUUID `json:"ids"`
}
