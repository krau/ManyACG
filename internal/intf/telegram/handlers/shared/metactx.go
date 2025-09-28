package shared

import "github.com/mymmrac/telego"

type HandlersMeta struct {
	ChannelChatID    telego.ChatID
	BotUsername      string
	ChannelAvailable bool
}

type HandlersMetaCtxKey struct{}

func MetaFromContext(ctx any) *HandlersMeta {
	if meta, ok := ctx.(*HandlersMeta); ok {
		return meta
	}
	return nil
}

func MustMetaFromContext(ctx any) *HandlersMeta {
	meta := MetaFromContext(ctx)
	if meta == nil {
		panic("HandlersMeta not found in context")
	}
	return meta
}
