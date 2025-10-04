package handlers

import (
	"context"

	"github.com/krau/ManyACG/internal/shared"
	"github.com/krau/ManyACG/service"

	"github.com/mymmrac/telego"
)

func CheckPermissionInGroup(ctx context.Context, message telego.Message, permissions ...shared.Permission) bool {
	chatID := message.Chat.ID
	if message.Chat.Type != telego.ChatTypeGroup && message.Chat.Type != telego.ChatTypeSupergroup {
		chatID = message.From.ID
	}
	if !service.CheckAdminPermissionByTgID(ctx, chatID, permissions...) {
		return service.CheckAdminPermissionByTgID(ctx, message.From.ID, permissions...)
	}
	return true
}

func CheckPermissionForQuery(ctx context.Context, query telego.CallbackQuery, permissions ...shared.Permission) bool {
	if !service.CheckAdminPermissionByTgID(ctx, query.From.ID, permissions...) &&
		!service.CheckAdminPermissionByTgID(ctx, query.Message.GetChat().ID, permissions...) {
		return false
	}
	return true
}
