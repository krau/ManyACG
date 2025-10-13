package handlers

import (
	"github.com/gofiber/fiber/v3"
	"github.com/krau/ManyACG/internal/infra/config/runtimecfg"
	"github.com/krau/ManyACG/internal/service"
)

func Register(router fiber.Router, serv *service.Service, cfg runtimecfg.RestConfig) {
	router.Get("/atom", GenerateAtomFeed)
	router.Get("/myip", MyIP(cfg))

	artworkGroup := router.Group("/artwork")
	artworkGroup.Get("/random", HandleRandomArtworks)
	artworkGroup.Post("/random", HandleRandomArtworks)
	artworkGroup.Get("/list", HandleListArtworks)
	artworkGroup.Post("/list", HandleListArtworks)
	artworkGroup.Get("/count", HandleCountArtwork)
	artworkGroup.Get("/fetch", HandleFetchArtwork)
	artworkGroup.Post("/fetch", HandleFetchArtwork)
	artworkGroup.Get("/:id", HandleGetArtworkByID)
}
