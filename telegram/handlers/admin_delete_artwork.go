package handlers

import (
	. "ManyACG/logger"
	"ManyACG/service"
	"ManyACG/storage"
	"ManyACG/telegram/utils"
	"ManyACG/types"
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegoutil"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func DeletePicture(ctx context.Context, bot *telego.Bot, message telego.Message) {
	if !CheckPermissionInGroup(ctx, message, types.PermissionDeleteArtwork) {
		utils.ReplyMessage(bot, message, "你没有删除图片的权限")
		return
	}
	var channelMessageID int
	cmd, _, args := telegoutil.ParseCommand(message.Text)
	if message.ReplyToMessage == nil {
		if len(args) == 0 {
			utils.ReplyMessage(bot, message, "用法:\n使用 /del 回复一条频道的图片消息, 或者提供频道消息ID.\n若使用 /delete 则删除整个作品")
			return
		}
		var err error
		channelMessageID, err = strconv.Atoi(args[0])
		if err != nil {
			utils.ReplyMessage(bot, message, "请不要输入奇怪的东西")
			return
		}
	} else {
		originChannel, ok := utils.GetMessageOriginChannelArtworkPost(ctx, bot, message)
		if !ok {
			utils.ReplyMessage(bot, message, "用法:\n使用 /del 回复一条频道的图片消息, 或者提供频道消息ID.\n若使用 /delete 则删除整个作品")
			return
		}
		channelMessageID = originChannel.MessageID
	}

	if channelMessageID == 0 {
		return
	}

	if cmd == "del" {
		picture, err := service.GetPictureByMessageID(ctx, channelMessageID)
		if err != nil {
			utils.ReplyMessage(bot, message, "获取图片信息失败: "+err.Error())
			return
		}
		if err := service.DeletePictureByMessageID(ctx, channelMessageID); err != nil {
			utils.ReplyMessage(bot, message, "从数据库中删除失败: "+err.Error())
			return
		}
		utils.ReplyMessage(bot, message, fmt.Sprintf("在数据库中已删除消息id为 %d 的图片", channelMessageID))
		bot.DeleteMessage(telegoutil.Delete(ChannelChatID, channelMessageID))

		if err := storage.DeleteAll(ctx, picture.StorageInfo); err != nil {
			Logger.Warnf("删除图片失败: %s", err)
			bot.SendMessage(telegoutil.Message(telegoutil.ID(message.From.ID), "在存储中删除图片文件失败: "+err.Error()))
		}
		return
	}
	artwork, err := service.GetArtworkByMessageID(ctx, channelMessageID)
	if err != nil {
		utils.ReplyMessage(bot, message, "获取作品信息失败: "+err.Error())
		return
	}
	if err := service.DeleteArtworkByURL(ctx, artwork.SourceURL); err != nil {
		utils.ReplyMessage(bot, message, "从数据库中删除失败: "+err.Error())
		return
	}
	utils.ReplyMessage(bot, message, fmt.Sprintf("在数据库中已删除消息id为 %d 的作品", channelMessageID))
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
			bot.SendMessage(telegoutil.Message(telegoutil.ID(message.From.ID), "删除图片失败: "+err.Error()))
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
