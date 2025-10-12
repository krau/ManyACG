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
	artworkGroup.Get("/random", RandomArtworks)
	artworkGroup.Get("/list", ListArtworks)
}
