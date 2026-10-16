package handlers

import (
	"github.com/gofiber/fiber/v3"
	"github.com/krau/ManyACG/internal/infra/config/runtimecfg"
)

func Register(router fiber.Router, cfg runtimecfg.RestConfig, app runtimecfg.AppConfig) {
	router.Get("/myip", MyIP(cfg))
	if app.IsMain() {
		artworkGroup := router.Group("/artwork")
		artworkGroup.Get("/fetch", HandleFetchArtwork)
		artworkGroup.Post("/fetch", HandleFetchArtwork)

		tgbotGroup := router.Group("/bot")
		tgbotGroup.Get("/send_artwork_info", HandleSendArtworkInfoByTelegramBot)
		tgbotGroup.Post("/send_artwork_info", HandleSendArtworkInfoByTelegramBot)
	}
	if !app.Capabilities.Has(runtimecfg.CapabilityRead) {
		return
	}

	router.Get("/atom", GenerateAtomFeed)

	artworkGroup := router.Group("/artwork")
	artworkGroup.Get("/random/preview", HandleRandomPreviewArtworks)
	artworkGroup.Get("/random", HandleRandomArtworks)
	artworkGroup.Post("/random", HandleRandomArtworks)
	artworkGroup.Get("/list", GetHandleListArtworks(cfg))
	artworkGroup.Post("/list", GetHandleListArtworks(cfg))
	artworkGroup.Get("/count", HandleCountArtwork)
	artworkGroup.Get("/:id", HandleGetArtworkByID)

	pictureGroup := router.Group("/picture")
	pictureGroup.Get("/file/:size/:id", NewSizedPictureFileHandler(app.IsMain()))
	pictureGroup.Get("/file/:id", HandleGetPictureFileByID)
	pictureGroup.Get("/random", HandleGetRandomPicture)

	artistGroup := router.Group("/artist")
	artistGroup.Get("/:id", HandleGetArtistByID)

	tagGroup := router.Group("/tag")
	tagGroup.Get("/random", HandleGetRandomTags)
}
