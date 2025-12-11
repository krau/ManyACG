package utils

import (
	"context"

	"github.com/krau/ManyACG/internal/interface/telegram/metautil"
	"github.com/krau/ManyACG/internal/model/entity"
	"github.com/krau/ManyACG/internal/service"
	"github.com/krau/ManyACG/internal/shared"
	"github.com/krau/ManyACG/pkg/log"
)

func UpdateCachedArtworkFileID(ctx context.Context,
	results []MediaGroupResultMessage,
	serv *service.Service,
	meta *metautil.MetaData,
	artwork *entity.CachedArtworkData) error {
	for _, msg := range results {
		if msg.UgoiraIndex >= 0 {
			if len(artwork.UgoiraMetas) <= msg.UgoiraIndex {
				log.Warnf("ugoira index out of range: %d/%d", msg.UgoiraIndex, len(artwork.UgoiraMetas))
				continue
			}
			artwork.UgoiraMetas[msg.UgoiraIndex].TelegramInfo.SetFileID(meta.BotID(), shared.TelegramMediaTypeVideo, msg.FileID)
		} else if msg.PictureIndex >= 0 {
			if len(artwork.Pictures) <= msg.PictureIndex {
				log.Warnf("picture index out of range: %d/%d", msg.PictureIndex, len(artwork.Pictures))
				continue
			}
			artwork.Pictures[msg.PictureIndex].TelegramInfo.SetFileID(meta.BotID(), shared.TelegramMediaTypePhoto, msg.FileID)
		}
	}
	return serv.UpdateCachedArtwork(ctx, artwork)
}
