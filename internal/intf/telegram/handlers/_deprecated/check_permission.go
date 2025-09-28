package handlers

import (
	"context"

	"github.com/krau/ManyACG/service"
	"github.com/krau/ManyACG/types"

	"github.com/mymmrac/telego"
)

func CheckPermissionInGroup(ctx context.Context, message telego.Message, permissions ...types.Permission) bool {
	chatID := message.Chat.ID
	if message.Chat.Type != telego.ChatTypeGroup && message.Chat.Type != telego.ChatTypeSupergroup {
		chatID = message.From.ID
	}
	if !service.CheckAdminPermission(ctx, chatID, permissions...) {
		return service.CheckAdminPermission(ctx, message.From.ID, permissions...)
	}
	return true
}

func CheckPermissionForQuery(ctx context.Context, query telego.CallbackQuery, permissions ...types.Permission) bool {
	if !service.CheckAdminPermission(ctx, query.From.ID, permissions...) &&
		!service.CheckAdminPermission(ctx, query.Message.GetChat().ID, permissions...) {
		return false
	}
	return true
}
