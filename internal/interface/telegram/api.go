package telegram

// import (
// 	"context"

// 	"github.com/krau/ManyACG/telegram/utils"
// 	"github.com/krau/ManyACG/types"

// 	"github.com/mymmrac/telego"
// )

// var (
// 	sendArtworkInfoCh = make(chan *sendArtworkInfoParams, 100)
// )

// type sendArtworkInfoParams struct {
// 	Ctx    context.Context
// 	Bot    *telego.Bot
// 	Params *utils.SendArtworkInfoParams
// }

// func SendArtworkInfo(ctx context.Context, bot *telego.Bot, params *utils.SendArtworkInfoParams) error {
// 	if bot == nil {
// 		bot = Bot
// 	}
// 	select {
// 	case sendArtworkInfoCh <- &sendArtworkInfoParams{
// 		Ctx:    ctx,
// 		Bot:    bot,
// 		Params: params,
// 	}:
// 		return nil
// 	default:
// 		return utils.SendArtworkInfo(ctx, bot, params)
// 	}
// }

// func PostAndCreateArtwork(ctx context.Context, artwork *types.Artwork, bot *telego.Bot, fromID int64, messageID int) error {
// 	if bot == nil {
// 		bot = Bot
// 	}
// 	return utils.PostAndCreateArtwork(ctx, artwork, bot, fromID, messageID)
// }

// type Telegram struct{}

// func NewTelegram() *Telegram {
// 	return &Telegram{}
// }

// func (t *Telegram) GetArtworkHTMLCaption(artwork *types.Artwork) string {
// 	return utils.GetArtworkHTMLCaption(artwork)
// }

// func (t *Telegram) Bot() *telego.Bot {
// 	return Bot
// }

// func (t *Telegram) ChannelChatID() telego.ChatID {
// 	return ChannelChatID
// }
