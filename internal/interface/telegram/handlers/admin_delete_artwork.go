package handlers

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/krau/ManyACG/internal/interface/telegram/handlers/utils"
	"github.com/krau/ManyACG/internal/shared"
	"github.com/krau/ManyACG/pkg/log"
	"github.com/krau/ManyACG/pkg/objectuuid"
	"github.com/krau/ManyACG/service"
	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegohandler"
	"github.com/mymmrac/telego/telegoutil"
	"github.com/samber/oops"
)

func DeleteArtwork(ctx *telegohandler.Context, message telego.Message) error {
	serv := service.FromContext(ctx)
	if !utils.CheckPermissionInGroup(ctx, serv, message, shared.PermissionDeleteArtwork) {
		return oops.Errorf("user %d has no permission to delete artwork", message.From.ID)
	}
	var sourceURL string

	if message.ReplyToMessage != nil {
		sourceURL = utils.FindSourceURLInMessage(serv, message.ReplyToMessage)
	} else {
		sourceURL = serv.FindSourceURL(message.Text)
	}
	helpText := `
[管理员] <b>使用 /delete 命令回复一条包含作品链接的消息, 或在参数中提供作品链接, 将删除该作品</b>

命令语法: /delete [作品链接]
`
	if sourceURL == "" {
		utils.ReplyMessageWithHTML(ctx, message, helpText)
		return nil
	}
	artwork, err := serv.GetArtworkByURL(ctx, sourceURL)
	if err != nil {
		utils.ReplyMessageWithHTML(ctx, message, "获取作品信息失败: "+err.Error())
		return nil
	}
	cmd, _, args := telegoutil.ParseCommand(message.Text)

	if cmd == "delete" {
		if err := serv.DeleteArtworkByURL(ctx, sourceURL); err != nil {
			utils.ReplyMessageWithHTML(ctx, message, "删除作品失败: "+err.Error())
			return nil
		}
		utils.ReplyMessageWithHTML(ctx, message, "在数据库中已删除该作品")
		for _, picture := range artwork.Pictures {
			if err := serv.StorageDeleteByInfo(ctx, picture.StorageInfo.Data()); err != nil {
				log.Errorf("删除图片失败: %s", err)
			}
		}
		return nil
	}

	// cmd == del, 删除artwork里的图片
	if len(args) == 0 {
		utils.ReplyMessage(ctx, message, "请提供要删除的图片序号, 从1开始, 以空格分隔多个序号\n"+helpText)
		return nil
	}

	var pictureIndexes []int
	for _, arg := range args {
		pictureIndex, err := strconv.Atoi(arg)
		if err != nil {
			utils.ReplyMessage(ctx, message, fmt.Sprintf("参数错误, 请指定要删除的图片序号 (从1开始)\nerror: %s", err))
			return nil
		}
		if pictureIndex <= 0 || pictureIndex > len(artwork.Pictures) {
			utils.ReplyMessage(ctx, message, "图片索引越界, 该作品共有 "+strconv.Itoa(len(artwork.Pictures))+" 张图片")
			return nil
		}
		pictureIndexes = append(pictureIndexes, pictureIndex-1) // pictureIndex - 1 , 转为数组索引值
	}

	for _, pictureIndex := range pictureIndexes {
		picture := artwork.Pictures[pictureIndex]
		if err := serv.DeletePictureByID(ctx, picture.ID); err != nil {
			utils.ReplyMessage(ctx, message, fmt.Sprintf("删除第 %d 张图片失败\nerror: %s", pictureIndex+1, err))
			return oops.Errorf("delete picture %d of artwork %s failed: %w", pictureIndex, artwork.ID.String(), err)
		}
		if err := serv.StorageDeleteByInfo(ctx, picture.StorageInfo.Data()); err != nil {
			log.Errorf("删除图片失败: %s", err)
		}
	}
	utils.ReplyMessage(ctx, message, "在数据库中已删除所选图片")
	return nil
}

func DeleteArtworkCallbackQuery(ctx *telegohandler.Context, query telego.CallbackQuery) error {
	serv := service.FromContext(ctx)
	if !utils.CheckPermissionForQuery(ctx, serv, query, shared.PermissionDeleteArtwork) {
		ctx.Bot().AnswerCallbackQuery(ctx, telegoutil.CallbackQuery(query.ID).WithText("你没有删除图片的权限").WithCacheTime(60).WithShowAlert())
		return nil
	}

	// delete_artwork artwork_id
	args := strings.Split(query.Data, " ")
	if len(args) != 2 {
		ctx.Bot().AnswerCallbackQuery(ctx, telegoutil.CallbackQuery(query.ID).WithText("参数错误").WithCacheTime(60).WithShowAlert())
		return nil
	}

	artworkID, err := objectuuid.FromObjectIDHex(args[1])
	if err != nil {
		ctx.Bot().AnswerCallbackQuery(ctx, telegoutil.CallbackQuery(query.ID).WithText("无效的ID").WithCacheTime(60).WithShowAlert())
		return nil
	}

	artwork, err := serv.GetArtworkByID(ctx, artworkID)
	if err != nil {
		ctx.Bot().AnswerCallbackQuery(ctx, telegoutil.CallbackQuery(query.ID).WithText("获取作品信息失败: "+err.Error()).WithCacheTime(60).WithShowAlert())
		return nil
	}

	if err := serv.DeleteArtworkByID(ctx, artworkID); err != nil {
		ctx.Bot().AnswerCallbackQuery(ctx, telegoutil.CallbackQuery(query.ID).WithText("从数据库中删除失败: "+err.Error()).WithCacheTime(60).WithShowAlert())
		return nil
	}

	ctx.Bot().AnswerCallbackQuery(ctx, telegoutil.CallbackQuery(query.ID).WithText("在数据库中已删除该作品").WithCacheTime(60))

	for _, picture := range artwork.Pictures {
		if err := serv.StorageDeleteByInfo(ctx, picture.StorageInfo.Data()); err != nil {
			log.Warnf("删除图片失败: %s", err)
		}
	}
	return nil
}
