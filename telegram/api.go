package telegram

import (
	"context"

	"github.com/krau/ManyACG/telegram/utils"
	"github.com/krau/ManyACG/types"

	"github.com/mymmrac/telego"
)

func SendArtworkInfo(ctx context.Context, bot *telego.Bot, params *utils.SendArtworkInfoParams) error {
	if bot == nil {
		bot = Bot
	}
	return utils.SendArtworkInfo(ctx, bot, params)
}

func PostAndCreateArtwork(ctx context.Context, artwork *types.Artwork, bot *telego.Bot, fromID int64, messageID int) error {
	if bot == nil {
		bot = Bot
	}
	return utils.PostAndCreateArtwork(ctx, artwork, bot, fromID, messageID)
}
