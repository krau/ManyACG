package handlers

import (
	"fmt"
	"slices"
	"strconv"

	"github.com/krau/ManyACG/internal/interface/telegram/handlers/utils"
	"github.com/krau/ManyACG/internal/service"
	"github.com/krau/ManyACG/internal/shared"
	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegohandler"
	"github.com/mymmrac/telego/telegoutil"
	"github.com/samber/oops"
)

func SetAdmin(ctx *telegohandler.Context, message telego.Message) error {
	serv := service.FromContext(ctx)
	if !serv.CheckAdminPermissionByTgID(ctx, message.From.ID, shared.PermissionSudo) {
		return nil
	}
	var userID int64
	var userIDStr string
	cmd, _, args := telegoutil.ParseCommand(message.Text)

	var supportedPermissionsText string
	for _, p := range shared.PermissionNames() {
		supportedPermissionsText += fmt.Sprintf("<code>%s</code>\n", p)
	}
	if message.ReplyToMessage == nil {
		if len(args) == 0 {
			utils.ReplyMessageWithHTML(ctx, message,
				fmt.Sprintf("请回复一名用户或提供ID, 并指定权限, 以空格分隔, 支持的权限:\n%s", supportedPermissionsText),
			)
			return nil
		}
		var err error
		userIDStr = args[0]
		userID, err = strconv.ParseInt(userIDStr, 10, 64)
		if err != nil {
			utils.ReplyMessage(ctx, message, "无效的用户ID")
			return nil
		}
	} else {
		if message.ReplyToMessage.SenderChat != nil {
			userID = message.ReplyToMessage.SenderChat.ID
		} else {
			userID = message.ReplyToMessage.From.ID
		}
	}
	if userID == message.From.ID {
		utils.ReplyMessage(ctx, message, "不能修改自己的权限")
		return nil
	}

	deladmin := cmd == "deladmin"
	if deladmin {
		if err := serv.DeleteAdminByTgID(ctx, userID); err != nil {
			utils.ReplyMessage(ctx, message, "删除管理员失败: "+err.Error())
			return oops.Wrapf(err, "delete admin failed")
		}
		utils.ReplyMessage(ctx, message, "操作成功")
		return nil
	}
	inputPermissions := make([]shared.Permission, 0)
	for _, arg := range args {
		for _, p := range shared.PermissionValues() {
			if p.String() == arg {
				inputPermissions = append(inputPermissions, p)
				break
			}
		}
	}
	if len(inputPermissions) == 0 {
		utils.ReplyMessageWithHTML(ctx, message,
			fmt.Sprintf("请指定权限, 以空格分隔, 支持的权限:\n%s", supportedPermissionsText),
		)
		return nil
	}
	if userID <= -1000000000000 && slices.Contains(inputPermissions, shared.PermissionSudo) {
		utils.ReplyMessage(ctx, message, "不能赋予群组超级管理员权限")
		return nil
	}

	if err := serv.CreateOrUpdateAdmin(ctx, userID, inputPermissions); err != nil {
		utils.ReplyMessage(ctx, message, "更新管理员权限失败: "+err.Error())
		return oops.Wrapf(err, "update admin permissions failed")
	}
	utils.ReplyMessage(ctx, message, "操作成功")
	return nil
}

func AddTagAlias(ctx *telegohandler.Context, message telego.Message) error {
	serv := service.FromContext(ctx)
	if !serv.CheckAdminPermissionByTgID(ctx, message.From.ID, shared.PermissionSudo) {
		utils.ReplyMessage(ctx, message, "你没有权限添加标签别名")
		return nil
	}
	_, _, args := utils.ParseCommandBy(message.Text, " ", "\"")
	if len(args) < 2 {
		utils.ReplyMessage(ctx, message, "请提供原标签名和需要添加的别名")
		return nil
	}
	tagName := args[0]
	tagAliases := args[1:]
	tag, err := serv.GetTagByName(ctx, tagName)
	if err != nil {
		utils.ReplyMessage(ctx, message, "获取标签失败: "+err.Error())
		return nil
	}
	if _, err := serv.AddTagAlias(ctx, tag.ID, tagAliases); err != nil {
		utils.ReplyMessage(ctx, message, "添加标签别名失败: "+err.Error())
		return nil
	}
	utils.ReplyMessage(ctx, message, "添加标签别名成功")
	return nil
}
