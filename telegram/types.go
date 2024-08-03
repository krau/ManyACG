package telegram

import (
	"github.com/mymmrac/telego"
)

type SendArtworkInfoParams struct {
	ChatID        *telego.ChatID
	SourceURL     string
	AppendCaption string
	Verify        bool
	HasPermission bool
	IgnoreDeleted bool
	ReplyParams   *telego.ReplyParameters
}
