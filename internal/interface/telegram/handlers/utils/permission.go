package utils

import (
	"context"

	"github.com/krau/ManyACG/internal/service"
	"github.com/krau/ManyACG/internal/shared"
	"github.com/mymmrac/telego"
)

func CheckPermissionInGroup(ctx context.Context, serv *service.Service, message telego.Message, permissions ...shared.Permission) bool {
	chatID := message.Chat.ID
	if message.Chat.Type != telego.ChatTypeGroup && message.Chat.Type != telego.ChatTypeSupergroup {
		chatID = message.From.ID
	}
	if !serv.CheckAdminPermissionByTgID(ctx, chatID, permissions...) {
		return serv.CheckAdminPermissionByTgID(ctx, message.From.ID, permissions...)
	}
	return true
}

func CheckPermissionForQuery(ctx context.Context, serv *service.Service, query telego.CallbackQuery, permissions ...shared.Permission) bool {
	if !serv.CheckAdminPermissionByTgID(ctx, query.From.ID, permissions...) &&
		!serv.CheckAdminPermissionByTgID(ctx, query.Message.GetChat().ID, permissions...) {
		return false
	}
	return true
}
