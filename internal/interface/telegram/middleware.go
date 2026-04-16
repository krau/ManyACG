package telegram

import (
	"strings"

	"github.com/krau/ManyACG/pkg/log"
	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegohandler"
)

func messageLogger(ctx *telegohandler.Context, update telego.Update) error {
	if update.Message != nil {
		chat := update.Message.Chat
		user := update.Message.From
		senderChat := update.Message.SenderChat
		if senderChat != nil {
			log.Debug("received message", "chat_title", chat.Title, "chat_id", chat.ID, "sender", senderChat.Title, "sender_username", senderChat.Username)
		} else {
			log.Debug("received message", "chat_title", chat.Title, "chat_id", chat.ID, "sender", user.FirstName+user.LastName, "sender_username", user.ID)
		}
	}
	return ctx.Next(update)
}

func updateLogFields(update telego.Update) []any {
	fields := []any{"update_type", telegramUpdateType(update)}

	if msg := update.Message; msg != nil {
		fields = append(fields,
			"chat_id", msg.Chat.ID,
			"chat_title", msg.Chat.Title,
			"chat_type", msg.Chat.Type,
			"message_id", msg.MessageID,
			"text", msg.Text,
		)
		if msg.From != nil {
			fields = append(fields,
				"user_id", msg.From.ID,
				"username", msg.From.Username,
				"user_name", strings.TrimSpace(msg.From.FirstName+" "+msg.From.LastName),
			)
		}
		if msg.SenderChat != nil {
			fields = append(fields,
				"sender_chat_id", msg.SenderChat.ID,
				"sender_chat_title", msg.SenderChat.Title,
				"sender_chat_username", msg.SenderChat.Username,
			)
		}
		return fields
	}

	if cq := update.CallbackQuery; cq != nil {
		fields = append(fields, "callback_id", cq.ID, "data", cq.Data)
		fields = append(fields,
			"user_id", cq.From.ID,
			"username", cq.From.Username,
			"user_name", strings.TrimSpace(cq.From.FirstName+" "+cq.From.LastName),
		)
		if msg := cq.Message; msg != nil && msg.IsAccessible() {
			fields = append(fields,
				"chat_id", msg.GetChat().ID,
				"message_id", msg.GetMessageID(),
			)
		}
		return fields
	}

	if iq := update.InlineQuery; iq != nil {
		fields = append(fields, "inline_query_id", iq.ID, "query", iq.Query)
		fields = append(fields,
			"user_id", iq.From.ID,
			"username", iq.From.Username,
			"user_name", strings.TrimSpace(iq.From.FirstName+" "+iq.From.LastName),
		)
		return fields
	}

	if post := update.ChannelPost; post != nil {
		fields = append(fields,
			"chat_id", post.Chat.ID,
			"chat_title", post.Chat.Title,
			"message_id", post.MessageID,
			"text", post.Text,
		)
		return fields
	}

	return fields
}

func telegramUpdateType(update telego.Update) string {
	switch {
	case update.Message != nil:
		return "message"
	case update.CallbackQuery != nil:
		return "callback_query"
	case update.InlineQuery != nil:
		return "inline_query"
	case update.ChannelPost != nil:
		return "channel_post"
	default:
		return "unknown"
	}
}
