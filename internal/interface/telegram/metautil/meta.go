package metautil

import (
	"context"

	"github.com/mymmrac/telego"
)

type MetaData struct {
	ChannelChatID    telego.ChatID
	BotUsername      string
	channelAvailable bool
}

func (m *MetaData) ChannelAvailable() bool {
	return m.channelAvailable
}

type MetaDataCtxKey struct{}

func NewMetaData(channelChatID telego.ChatID, botUsername string) *MetaData {
	return &MetaData{
		ChannelChatID:    channelChatID,
		BotUsername:      botUsername,
		channelAvailable: channelChatID.ID != 0 || channelChatID.Username != "",
	}
}

func FromContext(ctx context.Context) *MetaData {
	if meta, ok := ctx.Value(MetaDataCtxKey{}).(*MetaData); ok {
		return meta
	}
	return &MetaData{}
}

func WithContext(ctx context.Context, meta *MetaData) context.Context {
	return context.WithValue(ctx, MetaDataCtxKey{}, meta)
}
