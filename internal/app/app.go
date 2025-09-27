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
