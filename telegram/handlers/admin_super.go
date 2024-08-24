package handlers

import (
	"ManyACG/common"
	"ManyACG/service"
	"ManyACG/telegram/utils"
	"ManyACG/types"
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegoutil"
	"go.mongodb.org/mongo-driver/mongo"
)

func ProcessPicturesHashAndSize(ctx context.Context, bot *telego.Bot, message telego.Message) {
	userAdmin, err := service.GetAdminByUserID(ctx, message.From.ID)
	if err != nil {
		utils.ReplyMessage(bot, message, "获取管理员信息失败: "+err.Error())
		return
	}
	if userAdmin != nil && !userAdmin.SuperAdmin {
		utils.ReplyMessage(bot, message, "你没有权限处理图片信息")
		return
	}
	go service.ProcessPicturesHashAndSizeAndUpdate(context.TODO(), bot, &message)
	utils.ReplyMessage(bot, message, "开始处理了")
}

func ProcessPicturesStorage(ctx context.Context, bot *telego.Bot, message telego.Message) {
	userAdmin, err := service.GetAdminByUserID(ctx, message.From.ID)
	if err != nil {
		utils.ReplyMessage(bot, message, "获取管理员信息失败: "+err.Error())
		return
	}
	if userAdmin != nil && !userAdmin.SuperAdmin {
		utils.ReplyMessage(bot, message, "你没有权限处理图片的存储信息")
		return
	}
	go service.StoragePicturesRegularAndThumbAndUpdate(context.TODO(), bot, &message)
	utils.ReplyMessage(bot, message, "开始处理了")
}

func SetAdmin(ctx context.Context, bot *telego.Bot, message telego.Message) {
	userAdmin, err := service.GetAdminByUserID(ctx, message.From.ID)
	if err != nil {
		utils.ReplyMessage(bot, message, "获取管理员信息失败: "+err.Error())
		return
	}
	if userAdmin == nil || !userAdmin.SuperAdmin {
		utils.ReplyMessage(bot, message, "你没有权限设置管理员")
		return
	}
	var userID int64
	var userIDStr string
	_, _, args := telegoutil.ParseCommand(message.Text)
	var supportedPermissionsText string
	for _, p := range types.AllPermissions {
		supportedPermissionsText += "`" + string(p) + "`" + "\n"
	}
	if message.ReplyToMessage == nil {
		if len(args) == 0 {
			utils.ReplyMessageWithMarkdown(
				bot, message,
				fmt.Sprintf("请回复一名用户或提供ID\\, 并提供权限\\, 以空格分隔\n支持的权限\\:\n%v", supportedPermissionsText),
			)
			return
		}
		var err error
		userIDStr = args[0]
		userID, err = strconv.ParseInt(userIDStr, 10, 64)
		if err != nil {
			utils.ReplyMessage(bot, message, "请不要输入奇怪的东西")
			return
		}
	} else {
		if message.ReplyToMessage.SenderChat != nil {
			userID = message.ReplyToMessage.SenderChat.ID
		} else {
			userID = message.ReplyToMessage.From.ID
		}
	}

	inputPermissions := make([]types.Permission, len(args)-1)
	unsupportedPermissions := make([]string, 0)
	for i, arg := range args[1:] {
		for _, p := range types.AllPermissions {
			if string(p) == arg {
				inputPermissions[i] = p
				break
			}
		}
		if inputPermissions[i] == "" {
			unsupportedPermissions = append(unsupportedPermissions, arg)
		}
	}

	if len(unsupportedPermissions) > 0 {
		utils.ReplyMessageWithMarkdown(bot, message, common.EscapeMarkdown(fmt.Sprintf("权限不存在: %v\n支持的权限:\n", unsupportedPermissions))+supportedPermissionsText)
		return
	}

	isAdmin, err := service.IsAdmin(ctx, userID)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			isSuper := len(inputPermissions) == 0
			if strings.HasPrefix(userIDStr, "-100") && isSuper {
				utils.ReplyMessage(bot, message, "禁止赋予群组所有权限")
				return
			}
			err := service.CreateOrUpdateAdmin(ctx, userID, inputPermissions, message.From.ID, isSuper)
			if err != nil {
				utils.ReplyMessage(bot, message, "设置管理员失败: "+err.Error())
				return
			}
			utils.ReplyMessage(bot, message, "设置管理员成功")
			return
		}
		utils.ReplyMessage(bot, message, "获取管理员信息失败: "+err.Error())
		return
	}
	if isAdmin {
		if (len(args) == 0 && message.ReplyToMessage != nil) || (len(args) == 1 && message.ReplyToMessage == nil) {
			if err := service.DeleteAdmin(ctx, userID); err != nil {
				utils.ReplyMessage(bot, message, "删除管理员失败: "+err.Error())
				return
			}
			utils.ReplyMessage(bot, message, fmt.Sprintf("删除管理员成功: %d", userID))
			return
		}
		err := service.CreateOrUpdateAdmin(ctx, userID, inputPermissions, message.From.ID, false)
		if err != nil {
			utils.ReplyMessage(bot, message, "更新管理员失败: "+err.Error())
			return
		}
		utils.ReplyMessage(bot, message, "更新管理员成功")
		return
	}
}
