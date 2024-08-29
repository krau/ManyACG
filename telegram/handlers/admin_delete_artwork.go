package handlers

import (
	. "ManyACG/logger"
	"ManyACG/service"
	"ManyACG/sources"
	"ManyACG/storage"
	"ManyACG/telegram/utils"
	"ManyACG/types"
	"context"
	"strings"

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
	if sourceURL == "" {
		utils.ReplyMessage(bot, message, "请回复一条消息, 或者指定作品链接")
		return
	}

	artwork, err := service.GetArtworkByURL(ctx, sourceURL)
	if err != nil {
		utils.ReplyMessage(bot, message, "获取作品信息失败: "+err.Error())
		return
	}
	if err := service.DeleteArtworkByURL(ctx, sourceURL); err != nil {
		utils.ReplyMessage(bot, message, "删除作品失败: "+err.Error())
		return
	}
	utils.ReplyMessage(bot, message, "在数据库中已删除该作品")
	for _, picture := range artwork.Pictures {
		if err := storage.DeleteAll(ctx, picture.StorageInfo); err != nil {
			Logger.Errorf("删除图片失败: %s", err)
		}
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

	artworkMessageIDs := make([]int, len(artwork.Pictures))
	for _, picture := range artwork.Pictures {
		if picture.TelegramInfo == nil || picture.TelegramInfo.MessageID == 0 {
			continue
		}
		artworkMessageIDs = append(artworkMessageIDs, picture.TelegramInfo.MessageID)
	}
	if len(artworkMessageIDs) > 0 {
		bot.DeleteMessages(&telego.DeleteMessagesParams{
			ChatID:     ChannelChatID,
			MessageIDs: artworkMessageIDs,
		})
	}

	for _, picture := range artwork.Pictures {
		if err := storage.DeleteAll(ctx, picture.StorageInfo); err != nil {
			Logger.Warnf("删除图片失败: %s", err)
			bot.AnswerCallbackQuery(telegoutil.CallbackQuery(query.ID).WithText("从存储中删除图片失败: " + err.Error()))
		}
	}

}
