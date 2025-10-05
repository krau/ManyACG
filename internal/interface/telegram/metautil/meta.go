package metautil

import (
	"context"
	"fmt"
	"strings"

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

var contextKey = MetaDataCtxKey{}

func NewMetaData(channelChatID telego.ChatID, botUsername string) *MetaData {
	return &MetaData{
		ChannelChatID:    channelChatID,
		BotUsername:      botUsername,
		channelAvailable: channelChatID.ID != 0 || channelChatID.Username != "",
	}
}

func FromContext(ctx context.Context) *MetaData {
	if meta, ok := ctx.Value(contextKey).(*MetaData); ok {
		return meta
	}
	return &MetaData{}
}

func WithContext(ctx context.Context, meta *MetaData) context.Context {
	return context.WithValue(ctx, contextKey, meta)
}

func (m *MetaData) BotDeepLink(cmd string, params ...string) string {
	return fmt.Sprintf("https://t.me/%s/?start=%s_%s", m.BotUsername, cmd, strings.Join(params, "_"))
}

func (m *MetaData) ChannelMessageURL(messageID int) string {
	if m.ChannelChatID.Username != "" {
		return fmt.Sprintf("https://t.me/%s/%d", strings.TrimPrefix(m.ChannelChatID.String(), "@"), messageID)
	}
	return fmt.Sprintf("https://t.me/c/%s/%d", strings.TrimPrefix(m.ChannelChatID.String(), "-100"), messageID)
}