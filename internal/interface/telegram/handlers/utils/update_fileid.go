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
		switch msg.Type {
		case MediaResultTypePhoto:
			if len(artwork.Pictures) <= msg.Index {
				log.Warnf("picture index out of range: %d/%d", msg.Index, len(artwork.Pictures))
				continue
			}
			artwork.Pictures[msg.Index].TelegramInfo.SetFileID(meta.BotID(), shared.TelegramMediaTypePhoto, msg.FileID)
		case MediaResultTypeUgoira:
			if len(artwork.UgoiraMetas) <= msg.Index {
				log.Warnf("ugoira index out of range: %d/%d", msg.Index, len(artwork.UgoiraMetas))
				continue
			}
			artwork.UgoiraMetas[msg.Index].TelegramInfo.SetFileID(meta.BotID(), shared.TelegramMediaTypeVideo, msg.FileID)
		case MediaResultTypeVideo:
			if len(artwork.Videos) <= msg.Index {
				log.Warnf("video index out of range: %d/%d", msg.Index, len(artwork.Videos))
				continue
			}
			artwork.Videos[msg.Index].TelegramInfo.SetFileID(meta.BotID(), shared.TelegramMediaTypeVideo, msg.FileID)
		default:
			log.Errorf("unknown media result type: %d", msg.Type)
		}
	}
	return serv.UpdateCachedArtwork(ctx, artwork)
}
