package app

import "github.com/krau/ManyACG/internal/app/command"

type Application struct {
	Commands Commands
	Queries  Queries
}

type Commands struct {
	ArtworkCreate command.CreateArtworkHandler
}

type Queries struct{}

func NewApplication(commands Commands, queries Queries) *Application {
	return &Application{
		Commands: commands,
		Queries:  queries,
	}
}
