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
	ArtworkUpdate command.UpdateArtworkHandler
}

type Queries struct {
	ArtworkQuery  query.ArtworkQueryHandler
	ArtworkSearch query.ArtworkSearchQueryHandler
	AdminQuery    query.AdminQueryHandler
}

func NewApplication(commands Commands, queries Queries) *Application {
	return &Application{
		Commands: commands,
		Queries:  queries,
	}
}
