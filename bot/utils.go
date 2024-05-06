package bot

import (
	"ManyACG-Bot/service"
	"ManyACG-Bot/types"
	"context"

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
