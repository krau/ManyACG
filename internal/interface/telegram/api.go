package telegram

import (
	"context"

	"github.com/krau/ManyACG/internal/interface/telegram/handlers/utils"
	"github.com/krau/ManyACG/internal/model/entity"
	"github.com/krau/ManyACG/internal/service"
	"github.com/mymmrac/telego"
	"github.com/samber/oops"
)

func (b *BotApp) PostAndCreateArtwork(ctx context.Context, serv *service.Service, artwork *entity.CachedArtworkData) error {
	if err := utils.PostAndCreateArtwork(ctx, b.Bot(), serv, artwork, telego.ChatID{}, b.channelChatID, 0); err != nil {
		return oops.Wrapf(err, "posting and creating artwork %s", artwork.SourceURL)
	}
	return nil
}
