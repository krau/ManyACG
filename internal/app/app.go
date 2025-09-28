package app

import (
	"github.com/krau/ManyACG/internal/app/command"
	"github.com/krau/ManyACG/internal/app/query"
)

type Application struct {
	Commands Commands
	Queries  Queries
}

type Commands struct {
	ArtworkCreate command.CreateArtworkHandler
}

type Queries struct {
	ArtworkQuery  query.ArtworkQueryHandler
	ArtworkSearch query.ArtworkSearchQueryHandler
}

func NewApplication(commands Commands, queries Queries) *Application {
	return &Application{
		Commands: commands,
		Queries:  queries,
	}
}
