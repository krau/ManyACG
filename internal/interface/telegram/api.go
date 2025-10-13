package telegram

import (
	"context"

	"github.com/krau/ManyACG/internal/interface/telegram/handlers/utils"
	"github.com/krau/ManyACG/internal/model/entity"
	"github.com/krau/ManyACG/internal/service"
	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegoutil"
	"github.com/samber/oops"
)

func (b *BotApp) PostAndCreateArtwork(ctx context.Context, serv *service.Service, artwork *entity.CachedArtworkData) error {
	var adminId telego.ChatID
	if len(b.cfg.Admins) > 0 {
		adminId = telegoutil.ID(b.cfg.Admins[0])
	} else {
		adminIds, _ := serv.GetAdminUserIDs(ctx)
		if len(adminIds) > 0 {
			adminId = telegoutil.ID(adminIds[0])
		}
	}
	if err := utils.PostAndCreateArtwork(ctx, b.Bot(), serv, artwork, adminId, b.channelChatID, 0); err != nil {
		return oops.Wrapf(err, "posting and creating artwork %s", artwork.SourceURL)
	}
	return nil
}
