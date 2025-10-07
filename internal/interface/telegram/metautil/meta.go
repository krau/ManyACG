package metautil

import (
	"context"
	"fmt"
	"strings"

	"github.com/mymmrac/telego"
)

type MetaData struct {
	channelChatID    telego.ChatID
	botUsername      string
	siteUrl          string
	channelAvailable bool
}

func (m *MetaData) ChannelChatID() telego.ChatID {
	return m.channelChatID
}

func (m *MetaData) BotUsername() string {
	return m.botUsername
}

func (m *MetaData) ChannelAvailable() bool {
	return m.channelAvailable
}

func (m *MetaData) SiteURL() string {
	return m.siteUrl
}

type MetaDataCtxKey struct{}

var contextKey = MetaDataCtxKey{}

type Option func(*MetaData)

func WithSiteURL(url string) Option {
	return func(m *MetaData) {
		m.siteUrl = strings.TrimRight(url, "/")
	}
}

func NewMetaData(channelChatID telego.ChatID, botUsername string, opts ...Option) *MetaData {
	meta := &MetaData{
		channelChatID:    channelChatID,
		botUsername:      botUsername,
		channelAvailable: channelChatID.ID != 0 || channelChatID.Username != "",
	}
	for _, opt := range opts {
		opt(meta)
	}
	return meta
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
	return fmt.Sprintf("https://t.me/%s/?start=%s_%s", m.botUsername, cmd, strings.Join(params, "_"))
}

func (m *MetaData) ChannelMessageURL(messageID int) string {
	if m.channelChatID.Username != "" {
		return fmt.Sprintf("https://t.me/%s/%d", strings.TrimPrefix(m.channelChatID.String(), "@"), messageID)
	}
	return fmt.Sprintf("https://t.me/c/%s/%d", strings.TrimPrefix(m.channelChatID.String(), "-100"), messageID)
}
