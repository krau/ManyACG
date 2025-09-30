package handlers

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/krau/ManyACG/common"
	"github.com/krau/ManyACG/internal/shared"
	"github.com/krau/ManyACG/service"
	"github.com/krau/ManyACG/telegram/utils"
	"github.com/krau/ManyACG/types"

	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegohandler"
	"github.com/mymmrac/telego/telegoutil"
	"go.mongodb.org/mongo-driver/mongo"
)

func ProcessPicturesHashAndSize(ctx *telegohandler.Context, message telego.Message) error {
	userAdmin, err := service.GetAdminByTgID(ctx, message.From.ID)
	if err != nil {
		common.Logger.Errorf("获取管理员信息失败: %s", err)
		utils.ReplyMessage(ctx, ctx.Bot(), message, "获取管理员信息失败")
		return nil
	}
	if userAdmin != nil && !userAdmin.SuperAdmin {
		utils.ReplyMessage(ctx, ctx.Bot(), message, "你没有权限处理图片信息")
		return nil
	}
	go service.ProcessPicturesHashAndSizeAndUpdate(context.Background(), ctx.Bot(), &message)
	utils.ReplyMessage(ctx, ctx.Bot(), message, "开始处理了")
	return nil
}

// func ProcessPicturesStorage(ctx *telegohandler.Context, message telego.Message) error {
// 	userAdmin, err := service.GetAdminByUserID(ctx, message.From.ID)
// 	if err != nil {
// 		common.Logger.Errorf("获取管理员信息失败: %s", err)
// 		utils.ReplyMessage(ctx, ctx.Bot(), message, "获取管理员信息失败")
// 		return nil
// 	}
// 	if userAdmin != nil && !userAdmin.SuperAdmin {
// 		utils.ReplyMessage(ctx, ctx.Bot(), message, "你没有权限处理图片的存储信息")
// 		return nil
// 	}
// 	go service.StoragePicturesRegularAndThumbAndUpdate(ctx, ctx.Bot(), &message)
// 	utils.ReplyMessage(ctx, ctx.Bot(), message, "开始处理了")
// 	return nil
// }

func FixTwitterArtists(ctx *telegohandler.Context, message telego.Message) error {
	userAdmin, err := service.GetAdminByTgID(ctx, message.From.ID)
	if err != nil {
		common.Logger.Errorf("获取管理员信息失败: %s", err)
		utils.ReplyMessage(ctx, ctx.Bot(), message, "获取管理员信息失败")
		return nil
	}
	if userAdmin != nil && !userAdmin.SuperAdmin {
		utils.ReplyMessage(ctx, ctx.Bot(), message, "你没有权限修复Twitter作者信息")
		return nil
	}
	go service.FixTwitterArtists(ctx, ctx.Bot(), &message)
	utils.ReplyMessage(ctx, ctx.Bot(), message, "开始处理了")
	return nil
}

func SetAdmin(ctx *telegohandler.Context, message telego.Message) error {
	userAdmin, err := service.GetAdminByTgID(ctx, message.From.ID)
	if err != nil {
		utils.ReplyMessage(ctx, ctx.Bot(), message, "获取管理员信息失败: "+err.Error())
		return nil
	}
	if userAdmin == nil || !userAdmin.SuperAdmin {
		utils.ReplyMessage(ctx, ctx.Bot(), message, "你没有权限设置管理员")
		return nil
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
			utils.ReplyMessageWithMarkdown(ctx,
				ctx.Bot(), message,
				fmt.Sprintf("请回复一名用户或提供ID\\, 并提供权限\\, 以空格分隔\n支持的权限\\:\n%v", supportedPermissionsText),
			)
			return nil
		}
		var err error
		userIDStr = args[0]
		userID, err = strconv.ParseInt(userIDStr, 10, 64)
		if err != nil {
			utils.ReplyMessage(ctx, ctx.Bot(), message, "请不要输入奇怪的东西")
			return nil
		}
	} else {
		if message.ReplyToMessage.SenderChat != nil {
			userID = message.ReplyToMessage.SenderChat.ID
		} else {
			userID = message.ReplyToMessage.From.ID
		}
	}

	inputPermissions := make([]shared.Permission, len(args)-1)
	unsupportedPermissions := make([]string, 0)
	for i, arg := range args[1:] {
		for _, p := range shared.PermissionValues() {
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
		utils.ReplyMessageWithMarkdown(ctx, ctx.Bot(), message, common.EscapeMarkdown(fmt.Sprintf("权限不存在: %v\n支持的权限:\n", unsupportedPermissions))+supportedPermissionsText)
		return nil
	}

	isAdmin, err := service.IsAdminByTgID(ctx, userID)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			isSuper := len(inputPermissions) == 0
			if strings.HasPrefix(userIDStr, "-100") && isSuper {
				utils.ReplyMessage(ctx, ctx.Bot(), message, "禁止赋予群组所有权限")
				return nil
			}
			err := service.CreateOrUpdateAdmin(ctx, userID, inputPermissions, message.From.ID, isSuper)
			if err != nil {
				utils.ReplyMessage(ctx, ctx.Bot(), message, "设置管理员失败: "+err.Error())
				return nil
			}
			utils.ReplyMessage(ctx, ctx.Bot(), message, "设置管理员成功")
			return nil
		}
		utils.ReplyMessage(ctx, ctx.Bot(), message, "获取管理员信息失败: "+err.Error())
		return nil
	}
	if isAdmin {
		if (len(args) == 0 && message.ReplyToMessage != nil) || (len(args) == 1 && message.ReplyToMessage == nil) {
			if err := service.DeleteAdminByTgID(ctx, userID); err != nil {
				utils.ReplyMessage(ctx, ctx.Bot(), message, "删除管理员失败: "+err.Error())
				return nil
			}
			utils.ReplyMessage(ctx, ctx.Bot(), message, fmt.Sprintf("删除管理员成功: %d", userID))
			return nil
		}
		err := service.CreateOrUpdateAdmin(ctx, userID, inputPermissions, message.From.ID, false)
		if err != nil {
			utils.ReplyMessage(ctx, ctx.Bot(), message, "更新管理员失败: "+err.Error())
			return nil
		}
		utils.ReplyMessage(ctx, ctx.Bot(), message, "更新管理员成功")
		return nil
	}
	return nil
}

func AddTagAlias(ctx *telegohandler.Context, message telego.Message) error {
	userAdmin, err := service.GetAdminByTgID(ctx, message.From.ID)
	if err != nil {
		common.Logger.Errorf("获取管理员信息失败: %s", err)
		utils.ReplyMessage(ctx, ctx.Bot(), message, "获取管理员信息失败")
		return nil
	}
	if userAdmin != nil && !userAdmin.SuperAdmin {
		utils.ReplyMessage(ctx, ctx.Bot(), message, "你没有权限添加标签别名")
		return nil
	}
	_, _, args := utils.ParseCommandBy(message.Text, " ", "\"")
	if len(args) < 2 {
		utils.ReplyMessage(ctx, ctx.Bot(), message, "请提供原标签名和需要添加的别名")
		return nil
	}
	tagName := args[0]
	tagAliases := args[1:]
	tag, err := service.GetTagByName(ctx, tagName)
	if err != nil {
		utils.ReplyMessage(ctx, ctx.Bot(), message, "获取标签失败: "+err.Error())
		return nil
	}
	if _, err := service.AddTagAliasByID(ctx, tag.ID, tagAliases...); err != nil {
		utils.ReplyMessage(ctx, ctx.Bot(), message, "添加标签别名失败: "+err.Error())
		return nil
	}
	utils.ReplyMessage(ctx, ctx.Bot(), message, "添加标签别名成功")
	return nil
}

func AutoTagAllArtwork(ctx *telegohandler.Context, message telego.Message) error {
	userAdmin, err := service.GetAdminByTgID(ctx, message.From.ID)
	if err != nil {
		common.Logger.Errorf("获取管理员信息失败: %s", err)
		utils.ReplyMessage(ctx, ctx.Bot(), message, "获取管理员信息失败")
		return nil
	}
	if userAdmin != nil && !userAdmin.SuperAdmin {
		utils.ReplyMessage(ctx, ctx.Bot(), message, "你没有权限")
		return nil
	}
	go service.PredictAllArtworkTagsAndUpdate(ctx, ctx.Bot(), &message)
	utils.ReplyMessage(ctx, ctx.Bot(), message, "开始处理了")
	return nil
}
