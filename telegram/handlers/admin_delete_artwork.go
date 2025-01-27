package handlers

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/krau/ManyACG/common"
	"github.com/krau/ManyACG/service"
	"github.com/krau/ManyACG/sources"
	"github.com/krau/ManyACG/storage"
	"github.com/krau/ManyACG/telegram/utils"
	"github.com/krau/ManyACG/types"

	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegoutil"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func DeleteArtwork(ctx context.Context, bot *telego.Bot, message telego.Message) {
	if !CheckPermissionInGroup(ctx, message, types.PermissionDeleteArtwork) {
		utils.ReplyMessage(bot, message, "你没有删除作品的权限")
		return
	}
	var sourceURL string

	if message.ReplyToMessage != nil {
		sourceURL = utils.FindSourceURLForMessage(message.ReplyToMessage)
	} else {
		sourceURL = sources.FindSourceURL(message.Text)
	}
	helpText := fmt.Sprintf(`
[管理员] <b>使用 /delete 命令回复一条包含作品链接的消息, 或在参数中提供作品链接, 将删除该作品</b>

命令语法: %s
`, common.EscapeHTML("/delete [作品链接]"))
	if sourceURL == "" {
		utils.ReplyMessageWithHTML(bot, message, helpText)
		return
	}
	artwork, err := service.GetArtworkByURL(ctx, sourceURL)
	if err != nil {
		utils.ReplyMessage(bot, message, "获取作品信息失败: "+err.Error())
		return
	}
	cmd, _, args := telegoutil.ParseCommand(message.Text)

	if cmd == "delete" {
		if err := service.DeleteArtworkByURL(ctx, sourceURL); err != nil {
			utils.ReplyMessage(bot, message, "删除作品失败: "+err.Error())
			return
		}
		utils.ReplyMessage(bot, message, "在数据库中已删除该作品")
		for _, picture := range artwork.Pictures {
			if err := storage.DeleteAll(ctx, picture.StorageInfo); err != nil {
				common.Logger.Errorf("删除图片失败: %s", err)
			}
		}
		return
	}

	if len(args) == 0 {
		utils.ReplyMessage(bot, message, "请提供要删除的图片序号 (从1开始)")
		return
	}

	pictureIndex := 0
	pictureIndexStr := args[len(args)-1]
	pictureIndex, err = strconv.Atoi(pictureIndexStr)
	if err != nil {
		utils.ReplyMessage(bot, message, fmt.Sprintf("参数错误, 请指定要删除的图片序号 (从1开始)\nerror: %s", err))
		return
	}
	if pictureIndex <= 0 || pictureIndex > len(artwork.Pictures) {
		utils.ReplyMessage(bot, message, "请输入正确的图片序号, 从1开始")
		return
	}

	pictureIndex--

	picture := artwork.Pictures[pictureIndex]
	pictureID, err := primitive.ObjectIDFromHex(picture.ID)
	if err != nil {
		utils.ReplyMessage(bot, message, fmt.Sprintf("删除失败, 无效的ID\nerror: %s", err))
		return
	}
	if err := service.DeletePictureByID(ctx, pictureID); err != nil {
		utils.ReplyMessage(bot, message, fmt.Sprintf("删除失败\nerror: %s", err))
		return
	}
	utils.ReplyMessage(bot, message, "在数据库中已删除该图片")
	if err := storage.DeleteAll(ctx, picture.StorageInfo); err != nil {
		common.Logger.Errorf("删除图片失败: %s", err)
	}
}

func DeleteArtworkCallbackQuery(ctx context.Context, bot *telego.Bot, query telego.CallbackQuery) {
	if !CheckPermissionForQuery(ctx, query, types.PermissionDeleteArtwork) {
		bot.AnswerCallbackQuery(telegoutil.CallbackQuery(query.ID).WithText("你没有删除图片的权限").WithCacheTime(60).WithShowAlert())
		return
	}

	// delete_artwork artwork_id
	args := strings.Split(query.Data, " ")
	if len(args) != 2 {
		bot.AnswerCallbackQuery(telegoutil.CallbackQuery(query.ID).WithText("参数错误").WithCacheTime(60).WithShowAlert())
		return
	}

	artworkID, err := primitive.ObjectIDFromHex(args[1])
	if err != nil {
		bot.AnswerCallbackQuery(telegoutil.CallbackQuery(query.ID).WithText("无效的ID").WithCacheTime(60).WithShowAlert())
		return
	}

	artwork, err := service.GetArtworkByID(ctx, artworkID)
	if err != nil {
		bot.AnswerCallbackQuery(telegoutil.CallbackQuery(query.ID).WithText("获取作品信息失败: " + err.Error()).WithCacheTime(60).WithShowAlert())
		return
	}

	if err := service.DeleteArtworkByID(ctx, artworkID); err != nil {
		bot.AnswerCallbackQuery(telegoutil.CallbackQuery(query.ID).WithText("从数据库中删除失败: " + err.Error()).WithCacheTime(60).WithShowAlert())
		return
	}

	bot.AnswerCallbackQuery(telegoutil.CallbackQuery(query.ID).WithText("在数据库中已删除该作品").WithCacheTime(60))

	for _, picture := range artwork.Pictures {
		if err := storage.DeleteAll(ctx, picture.StorageInfo); err != nil {
			common.Logger.Warnf("删除图片失败: %s", err)
			bot.AnswerCallbackQuery(telegoutil.CallbackQuery(query.ID).WithText("从存储中删除图片失败: " + err.Error()))
		}
	}
}
