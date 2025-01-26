package types

import (
	"context"

	"github.com/mymmrac/telego"
)

type Service interface {
	GetArtworkByURL(ctx context.Context, url string, opts ...*AdapterOption) (*Artwork, error)
}

type TelegramService interface {
	GetArtworkHTMLCaption(artwork *Artwork) string
	Bot() *telego.Bot
	ChannelChatID() telego.ChatID
}
