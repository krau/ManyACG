package handlers

import (
	"context"

	"github.com/gofiber/fiber/v3"
	"github.com/krau/ManyACG/internal/infra/config/runtimecfg"
	"github.com/krau/ManyACG/internal/interface/rest/common"
	"github.com/krau/ManyACG/internal/model/entity"
	"github.com/krau/ManyACG/internal/shared"
	"github.com/krau/ManyACG/pkg/log"
	"gorm.io/datatypes"
)

func Register(router fiber.Router, cfg runtimecfg.RestConfig, writer common.WriteService) {
	router.Get("/atom", GenerateAtomFeed)
	router.Get("/myip", MyIP(cfg))

	artworkGroup := router.Group("/artwork")
	artworkGroup.Get("/random/preview", HandleRandomPreviewArtworks)
	artworkGroup.Get("/random", HandleRandomArtworks)
	artworkGroup.Post("/random", HandleRandomArtworks)
	artworkGroup.Get("/list", GetHandleListArtworks(cfg))
	artworkGroup.Post("/list", GetHandleListArtworks(cfg))
	artworkGroup.Get("/count", HandleCountArtwork)
	if writer != nil {
		artworkGroup.Get("/fetch", HandleFetchArtwork)
		artworkGroup.Post("/fetch", HandleFetchArtwork)
	}
	artworkGroup.Get("/:id", HandleGetArtworkByID)

	var pictureFileHandler fiber.Handler
	if writer != nil {
		pictureFileHandler = NewSizedPictureFileHandler(newPictureMIMEObserver(writer))
	} else {
		pictureFileHandler = NewSizedPictureFileHandler()
	}
	pictureGroup := router.Group("/picture")
	pictureGroup.Get("/file/:size/:id", pictureFileHandler)
	pictureGroup.Get("/file/:id", HandleGetPictureFileByID)
	pictureGroup.Get("/random", HandleGetRandomPicture)

	artistGroup := router.Group("/artist")
	artistGroup.Get("/:id", HandleGetArtistByID)

	tagGroup := router.Group("/tag")
	tagGroup.Get("/random", HandleGetRandomTags)

	if writer != nil {
		tgbotGroup := router.Group("/bot")
		tgbotGroup.Get("/send_artwork_info", HandleSendArtworkInfoByTelegramBot)
		tgbotGroup.Post("/send_artwork_info", HandleSendArtworkInfoByTelegramBot)
	}
}

func newPictureMIMEObserver(serv common.WriteService) PictureMIMEObserver {
	return func(picture *entity.Picture, size string, detail shared.StorageDetail) {
		updatedPicture := *picture
		storageInfo := picture.StorageInfo.Data()
		switch size {
		case "thumb":
			storageInfo.Thumb = &detail
		case "regular":
			storageInfo.Regular = &detail
		case "original":
			storageInfo.Original = &detail
		}
		updatedPicture.StorageInfo = datatypes.NewJSONType(storageInfo)
		go func() {
			if err := serv.SavePicture(context.Background(), &updatedPicture); err != nil {
				log.Errorf("failed to save picture mime: %v", err)
			}
		}()
	}
}
