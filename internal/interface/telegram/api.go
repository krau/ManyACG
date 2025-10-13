package telegram

import (
	"context"

	"github.com/krau/ManyACG/internal/interface/telegram/handlers/utils"
	"github.com/krau/ManyACG/internal/model/entity"
	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegoutil"
	"github.com/samber/oops"
)

func (b *BotApp) PostAndCreateArtwork(ctx context.Context, artwork *entity.CachedArtworkData) error {
	var adminId telego.ChatID
	if len(b.cfg.Admins) > 0 {
		adminId = telegoutil.ID(b.cfg.Admins[0])
	} else {
		adminIds, _ := b.serv.GetAdminUserIDs(ctx)
		if len(adminIds) > 0 {
			adminId = telegoutil.ID(adminIds[0])
		}
	}
	if err := utils.PostAndCreateArtwork(ctx, b.Bot(), b.serv, artwork, adminId, b.meta.ChannelChatID(), 0); err != nil {
		return oops.Wrapf(err, "posting and creating artwork %s", artwork.SourceURL)
	}
	return nil
}

func (b *BotApp) SendArtworkInfo(ctx context.Context, sourceUrl string, chatID int64, appendCaption string) error {
	return utils.SendArtworkInfo(ctx, b.bot, b.meta, b.serv, sourceUrl, telegoutil.ID(chatID), utils.SendArtworkInfoOptions{AppendCaption: appendCaption, HasPermission: true})
}
